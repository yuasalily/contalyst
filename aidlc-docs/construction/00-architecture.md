# Contalyst — Architecture

How the code is organized and how data flows through it. Pairs with
[01-implementation-status.md](./01-implementation-status.md) (what is built) and
[02-developer-guide.md](./02-developer-guide.md) (how to build/extend).

---

## 1. Module & package layout

Go module: `github.com/yuasalily/contalyst` (Go 1.26+).

```
contalyst/
├── main.go                      # entry point: build client, run the Bubble Tea program
├── internal/
│   ├── dockerx/                 # the ONLY package that imports the Docker SDK
│   │   ├── client.go            #   client + connection + container list/lifecycle/inspect
│   │   ├── logs.go              #   log streaming (TTY/non-TTY demux)
│   │   ├── stats.go             #   stats streaming + CPU%/mem derivation
│   │   ├── resources.go         #   images / volumes / networks list+remove+prune
│   │   ├── compose.go           #   [v2/U9] compose project grouping + `docker compose` shell-out
│   │   ├── context.go           #   [v2/U11] docker context enumeration + per-host client
│   │   ├── maintenance.go       #   [v2/U12] image history + disk-usage + per-category prune
│   │   └── util.go              #   JSON pretty-print helper
│   └── ui/                      # the Bubble Tea application
│       ├── app.go               #   root model, Update routing, layout, View composition
│       ├── messages.go          #   tea.Msg types + commands (async daemon calls, stream readers)
│       ├── keys.go              #   key bindings (bubbles/key)
│       ├── styles.go            #   lipgloss styles derived from a theme
│       ├── list.go              #   custom colorful table component (cursor, scroll, widths, mark gutter)
│       ├── rows.go              #   per-resource-kind column/row builders (incl. compose) + filter
│       ├── update_list.go       #   list-view key handling + container/compose/bulk/resource actions
│       ├── detail.go            #   logs + live-stats split view + inspect view
│       ├── overlays.go          #   filter, command palette, confirm, help, context/oplog/prune, header, hints
│       ├── context.go           #   [v2/U11] host/context switcher overlay + runtime client swap
│       ├── maintenance.go       #   [v2/U12] operation log + prune dashboard + image-layer view
│       ├── exec.go              #   exec into a shell via tea.ExecProcess
│       └── theme/
│           └── theme.go         #   color palettes + state→color/glyph mapping
└── aidlc-docs/                  # inception + construction documentation
```

### Dependency direction

```
main.go → internal/ui → internal/dockerx → github.com/docker/docker
                      ↘ internal/ui/theme
```

`internal/ui` never imports the Docker SDK directly; it speaks only to
`dockerx`'s domain types (`dockerx.Container`, `.Image`, `.Volume`, `.Network`,
`.Stats`, `.LogLine`). This is **CR-2 / inception NFR-M1**: it keeps the SDK
swappable (Docker is migrating `github.com/docker/docker` → `github.com/moby/moby`)
and keeps the UI testable without a daemon.

---

## 2. The Docker layer (`internal/dockerx`)

A thin wrapper that converts SDK types into small UI-facing structs and hides
SDK quirks.

- **Connection** — `NewClient()` uses `client.FromEnv` +
  `client.WithAPIVersionNegotiation()` so one binary works across Engine
  versions (inception R2). `Ping()` returns a human-actionable error.
- **Lists** — `Containers/Images/Volumes/Networks(ctx)` return sorted slices of
  domain structs; ports are pre-formatted, names de-slashed, digests shortened.
- **Lifecycle** — `Start/Stop/Restart/Pause/Unpause/Kill/Remove`, plus
  `Remove{Image,Volume,Network}` and `Prune{Containers,Images,Volumes,Networks}`.
- **Streaming** (the subtle part):
  - `LogStream(ctx, id, timestamps)` returns `<-chan LogLine`. It inspects the
    container's TTY flag and decodes accordingly: **non-TTY** streams carry an
    8-byte frame header and are demuxed through `stdcopy.StdCopy` via an
    `io.Pipe`; **TTY** streams are read raw (inception R1 — the classic Docker
    log-corruption bug). Cancel `ctx` to stop; the channel then closes.
  - `StatsStream(ctx, id)` returns `<-chan Stats`. CPU% is computed from the
    **delta** between `cpu_stats` and `precpu_stats` × online CPUs (inception
    FR-C11). On cgroup v2, `inactive_file` is subtracted from memory usage to
    match `docker stats`.

Both streams own a goroutine that exits when `ctx` is cancelled — no leaks
(inception NFR-R1).

---

## 3. The UI (`internal/ui`) — The Elm Architecture

Bubble Tea is Model–Update–View. There is **one** model (`ui.model` in
`app.go`); there are no nested `tea.Model`s. Sub-views are plain helper methods
on that model. This keeps all state in one struct and avoids message-routing
boilerplate (**CR-3**).

### 3.1 State

`model` holds:
- connection state (`ready`, `serverVer`, `fatalErr`),
- the active **state** (`stateList` / `stateDetail` / `stateInspect`) and active
  **overlay** (`ovNone` / `ovFilter` / `ovCommand` / `ovConfirm` / `ovHelp`),
- the active **resourceKind** (containers/images/volumes/networks) and the
  cached data slice for each,
- the `list` component, the `detail` and `inspect` sub-states,
- input components (`filterInput`, `cmdInput`), the `confirm` dialog state,
- the current `theme.Theme` + derived `styles`, and the transient `toast`.

`state` vs `overlay` are orthogonal: `state` is the main screen; `overlay` is a
modal layer on top (an input bar at the bottom, or a centered box).

### 3.2 Update routing (`app.go`)

`Update` first handles **data/lifecycle messages** (window size, connection,
ticks, list payloads, action results, stream events). Key presses are delegated
to `handleKey`, which routes by precedence:

```
fatalErr?  → only quit works
overlay?   → updateFilter / updateCommand / updateConfirm / help-dismiss
else state → updateDetail / updateInspect / updateList
```

### 3.3 View composition (`app.go View()`)

```
┌ headerView ┐  app name › breadcrumbs … docker <ver> · <theme>   (1 line)
│ body       │  lst.view  | detailView | inspectView
│ toastView  │  ✓/✕ message                                       (0–1 line)
│ bottomView │  hint bar (2 lines) | filter bar | command bar
└────────────┘
```

`headerHeight`, the hint height, and an optional toast line are subtracted from
the terminal height in `recomputeLayout()` to size the body, the list, the log
viewport, and the inspect viewport. Centered overlays (help, confirm) are
composited last via `overlayCenter` (a `lipgloss.Place` that replaces the screen
— a deliberate modal focus).

---

## 4. Asynchronous work & streaming pattern (`messages.go`)

Bubble Tea forbids blocking in `Update`. All daemon I/O is a `tea.Cmd` that runs
on a goroutine and returns a `tea.Msg`.

- **One-shot** calls (list, action, inspect) → a command that returns a single
  message (`containersMsg`, `actionDoneMsg`, `inspectMsg`, …).
- **Periodic refresh** — `tickCmd()` (`tea.Tick`, 1.5s) drives `loadCmd()` while
  in `stateList`. In detail view the tick keeps firing but does **not** reload,
  so it never disturbs a log stream.
- **Continuous streams** use the **start→wait→re-issue** idiom (**CR-4**):
  1. On entering detail, `enterDetail` creates a cancelable `context` (stored on
     the model) and issues `startLogCmd` / `startStatsCmd`.
  2. Those return `logStartedMsg{ch}` / `statsStartedMsg{ch}`; `Update` stores
     the channel and issues `waitLogCmd(ch)` / `waitStatsCmd(ch)`.
  3. Each `waitXCmd` blocks on one channel receive and returns one
     `logLineMsg` / `statsMsg`; `Update` applies it and **re-issues the same
     wait command**, pumping the stream one item at a time without blocking the
     event loop.
  4. Leaving detail calls `teardownStreams`, which cancels the contexts; the
     `dockerx` goroutines exit and close the channels (→ `logClosedMsg` /
     `statsClosedMsg`, which are no-ops).

---

## 5. The list component (`list.go` + `rows.go`)

A **custom** table renderer rather than `bubbles/table`. Reason (**CR-5**):
`bubbles/table` renders rows as plain strings and can't color individual cells,
but per-cell color — a green/red state glyph, cyan ports — is the core of the
"colorful" requirement. `list.go` owns cursor movement, vertical scrolling, and
column-width allocation (fixed + flexible columns sharing leftover width). It
renders unselected rows with per-cell color and the selected row with a
full-width background fill (plain text under the fill to avoid nested ANSI).

`rows.go` is the only kind-aware code: `rebuildList()` dispatches to
`buildContainers/Images/Volumes/Networks`, each defining that kind's columns and
mapping cached data to colored `listRow`s, applying the live fuzzy `filter`.
Adding a resource kind touches `rows.go`, `messages.go`, and `overlays.go` only
(recipe in [02-developer-guide.md](./02-developer-guide.md)).

---

## 6. Theming (`theme/theme.go` + `styles.go`)

`theme.Theme` is a palette of named `lipgloss.Color`s plus a state→color map and
a state→glyph map. Three themes ship (Catalyst/Aurora/Mono); `theme.Next` cycles
them. `styles.go` turns a `Theme` into a `styles` struct of pre-built
`lipgloss.Style`s. Switching theme = `applyTheme` rebuilds `styles` and the list
(`cycleTheme` in `app.go`, shared by the `T` key and `:theme`). Colors are
truecolor hex; termenv degrades them for 256/16-color terminals (inception
NFR-U3), and state is always shown by glyph too, so the UI is legible without
color.

---

## 7. Construction decisions

These are the implementation-level decisions (the inception `DR-x` decisions
cover product/scope). Recorded here so the rationale survives.

| ID | Decision | Rationale |
|---|---|---|
| CR-1 | Build on **Bubble Tea v1** (bubbletea v1.3.10, bubbles v1.0.0, lipgloss v1.1.0) | v1 is the mature line all `bubbles` components target; v2 was still stabilizing. Lip Gloss v2 also removes `AdaptiveColor` |
| CR-2 | Isolate the SDK in `dockerx` behind domain types | NFR-M1; eases the coming `docker/docker`→`moby/moby` move; makes the UI testable without a daemon |
| CR-3 | Single root model, sub-views as methods (no nested `tea.Model`) | Docker's UI is small; one state struct is simpler than message-routing between child models |
| CR-4 | start→wait→re-issue command idiom for streams; cancelable context per stream | Non-blocking event loop; guaranteed goroutine teardown (NFR-R1) |
| CR-5 | Custom list renderer instead of `bubbles/table` | Per-cell color is essential to the colorful brief and unsupported by `bubbles/table` |
| CR-6 | `exec` shells out to `docker exec -it` via `tea.ExecProcess` | Gives a correct interactive PTY for free; the docker CLI is virtually always present alongside the daemon (same approach as lazydocker) |
| CR-7 | Periodic polling (`tea.Tick` 1.5s) for live list updates | Robust and simple vs. subscribing to the Docker event stream; revisit if event-driven refresh is wanted |
| CR-8 | **[v2/U9]** Compose projects are *derived* from the live container list (`com.docker.compose.*` labels), not from parsing compose files | No extra state or file parsing; the project view is always consistent with what is actually running. `dockerx.ComposeProjects` is a pure function (unit-tested) |
| CR-9 | **[v2/U9]** Compose lifecycle ops shell out to `docker compose` (inception DR-5), passing the project's `config_files`/`working_dir` labels through `-f`/`--project-directory` | The SDK has no compose API; delegating matches `docker compose` semantics incl. `depends_on` ordering for free. Parity with CR-6. Availability is probed once (`composeAvailCmd`); the view degrades read-only when absent (R9) |
| CR-10 | **[v2/U11]** Contexts come from `docker context ls` (CLI), and switching rebuilds the `dockerx.Client` via `NewClientForHost` | Follows the Docker-standard context mechanism (DR-6) without re-implementing the on-disk format; client swap stays inside `dockerx` (NFR-M1). Streams are torn down and view state reset on switch (R11) |
| CR-11 | **[v2/U12]** The operation log is recorded centrally in `Update` on every `actionDoneMsg` | All one-shot/bulk/compose/prune actions already funnel through `actionDoneMsg`, so a single hook captures them uniformly — no threading through call sites (FR-OL1). Session-only ring buffer (OQ-6) |
| CR-12 | **[v2/U10]** Bulk actions fan out with one goroutine per target and aggregate the outcome (`bulkAction`); partial failure is tolerated | Matches the non-blocking command model; reports `N ok / M failed` rather than aborting on first error (FR-B5 / R10). The mark gutter is a width-reserving column on the custom list renderer (CR-5) |

---

## 8. Where things live (quick index)

| To change… | Edit |
|---|---|
| A Docker API call / data shape | `internal/dockerx/*` |
| A key binding | `keys.go` (+ handler in `update_list.go` / `detail.go`) |
| Colors / a new theme | `theme/theme.go` (+ `styles.go` if a new style is needed) |
| List columns / row colors | `rows.go` |
| The hint bar / help text | `overlays.go` (`hintView`, `helpBox`) |
| A command-palette verb | `overlays.go` (`runCommand`, `commandNames`) |
| Layout / sizing | `recomputeLayout` in `app.go` |
| The detail (logs+stats) view | `detail.go` |
| Compose grouping / ops | `dockerx/compose.go` (+ `rows.go` `buildCompose`, `update_list.go` `updateComposeActions`) |
| Bulk/multi-select behaviour | `update_list.go` (`updateBulkActions`, `toggleMarkAll`) + `list.go` (mark gutter) |
| Context/host switching | `dockerx/context.go` + `ui/context.go` |
| Op log / prune dashboard / layer view | `ui/maintenance.go` (+ `dockerx/maintenance.go`) |
