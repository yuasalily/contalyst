# Contalyst — Developer Guide

Everything needed to build, run, test, and extend Contalyst. Architecture
context is in [00-architecture.md](./00-architecture.md).

---

## 1. Prerequisites

- **Go 1.26+**
- A reachable **Docker daemon** (to run the app and the smoke test). The app
  connects via the standard environment — `DOCKER_HOST` or the default socket —
  and negotiates the API version, so a range of Engine versions works.
- The **`docker` CLI** on `PATH` (only the `exec` feature shells out to it).

## 2. Common tasks (Makefile)

```sh
make build            # build ./contalyst with version info injected
make run              # go run .
make test             # unit + functional tests (-race, no daemon)
make test-unit        # only unit tests       (go test -run TestUnit)
make test-functional  # only functional tests (go test -run TestFunctional)
make test-e2e         # E2E tests against the live daemon (-tags e2e)
make cover            # coverage report (unit + functional)
make lint             # gofmt check + go vet (the CI lint gate)
make snapshot         # local cross-platform release build (GoReleaser, no publish)
make install          # go install . (to GOBIN)
make help             # list all targets
```

A clean checkout builds to a single self-contained binary (no external runtime).
`make build`/`install` inject `version`/`commit`/`date` via `-ldflags` (see
`contalyst --version`).

## 3. Dependency & license inventory

Direct dependencies and the versions in use:

| Module | Version | License | Role |
|---|---|---|---|
| `github.com/charmbracelet/bubbletea` | v1.3.10 | MIT | TUI framework |
| `github.com/charmbracelet/bubbles` | v1.0.0 | MIT | `textinput`, `viewport`, `key` |
| `github.com/charmbracelet/lipgloss` | v1.1.0 | MIT | styling / layout |
| `github.com/docker/docker` | v28.5.2+incompatible | Apache-2.0 | Docker Engine SDK |
| `github.com/mattn/go-runewidth` | v0.0.24 | MIT | display-width-correct padding/truncation |

> **v2 added no new dependencies.** Compose and context support shell out to the
> `docker` / `docker compose` CLIs (CR-9/CR-10, same approach as `exec`/CR-6);
> the layer/disk-usage/history reads use the existing Docker SDK. Bulk fan-out
> uses the stdlib `sync` package. The dependency/license table above is
> unchanged by v2.

**License policy (inception DR-1):** the user specified Apache-2.0 for libraries
but explicitly named Bubble Tea, which (with Lip Gloss/Bubbles) is MIT. Resolved:
the Charm ecosystem (MIT) is accepted as part of the Bubble Tea choice; every
non-Charm dependency is Apache-2.0 (the Docker SDK); no copyleft anywhere. MIT is
permissive and freely combinable with Apache-2.0 in one binary. Full attribution
is in the repo-root [`NOTICE`](../../NOTICE); the project itself is Apache-2.0
([`LICENSE`](../../LICENSE)).

> **SDK migration note (inception DR-2):** Docker Engine v29+ deprecates
> `github.com/docker/docker` in favor of independently-versioned
> `github.com/moby/moby/client` (license unchanged). All SDK use is confined to
> `internal/dockerx`, so the move is localized to that package.

## 4. Manual verification (TTY)

Automated tests cover logic without a TTY. To eyeball the real UI, start a few
containers and run it:

```sh
docker run -d --name demo-web -p 8088:80 nginx:alpine
docker run -d --name demo-log busybox sh -c 'i=0; while true; do echo "tick $i"; i=$((i+1)); sleep 1; done'
./contalyst
```

Check: colored state glyphs in the list; `⏎` opens logs + live stats; `f`
pauses/resumes follow; `s`/`r` change state; `d` shows a confirm dialog focused
on **Cancel**; `e` drops into a shell and returns; `:images` switches view; `/`
filters; `T` cycles themes; `?` shows help.

---

## 5. Extension recipes

### 5.1 Add a key binding / action

1. Add a `key.Binding` field to `keyMap` in **`keys.go`** and initialize it in
   `defaultKeys()` (avoid terminal-reserved chords `Ctrl-D/R/S` — inception R4).
2. Handle it where appropriate: list/container actions in
   **`update_list.go`** (`updateContainerActions` / `updateResourceActions`),
   detail in **`detail.go`** (`updateDetail`).
3. For a daemon call, add a method to **`dockerx`** and wrap it with `action()`
   from **`messages.go`** so the result becomes a success/error toast.
4. Add the key to the hint bar (`hintView`) and/or help (`helpBox`) in
   **`overlays.go`** so it stays discoverable.

### 5.2 Add a destructive action safely

Use `openConfirm(title, body, danger, onYes)` (in `overlays.go`). It opens the
confirm overlay with the cursor on **Cancel** (inception R5). Pass the actual
mutation as the `onYes` command (built with `action(...)`). Example in
`updateContainerActions` (`Delete`/`Kill` cases).

### 5.3 Add a new resource kind

(e.g. "contexts"). Touch three files:
1. **`messages.go`** — add a `resourceKind` const + `String()` case, a payload
   message type (`fooMsg`), and a branch in `loadCmd`.
2. **`dockerx`** — add the list/remove methods + a domain struct.
3. **`rows.go`** — add a `buildFoo()` (columns + colored rows + filter) and a
   `rebuildList()` case. Add the kind to **`overlays.go`** `runCommand` +
   `commandNames` so `:foo` switches to it.

### 5.4 Add a theme

In **`theme/theme.go`**, define a new `Theme` var (palette + it inherits the
state mapping) and add it to the `themes` slice. `theme.Next` and `T`/`:theme`
pick it up automatically. If it needs a brand-new styled element, add a field to
`styles` in **`styles.go`**.

### 5.5 Add a command-palette verb

In **`overlays.go`**: add the word to `commandNames` (drives tab-autosuggest) and
a `case` in `runCommand`.

---

## 5a. Testing strategy (three tiers)

Tests are layered so the fast tiers need no Docker and only the top tier does.
Tiers are selected by the test-name prefix and a build tag.

| Tier | Selector | Needs Docker? | What it covers | Files |
|---|---|---|---|---|
| **Unit** | `go test -run TestUnit ./...` | No | Pure functions: port/size/time formatting, CPU%/mem math, fuzzy match, command suggest, column widths, hint packing, theme mapping | `dockerx/unit_test.go`, `ui/unit_test.go`, `ui/theme/theme_test.go` |
| **Functional** | `go test -run TestFunctional ./...` | No | The whole UI model driven through `Update`/`View` with synthetic messages: list/detail/help/confirm/filter/command/theme + tiny-terminal | `ui/functional_test.go` |
| **E2E** | `go test -tags e2e -run TestE2E ./...` | **Yes** | Real daemon: `dockerx` lifecycle/log-demux/stats/inspect, and the **compiled program** driven via `teatest` (real client, real container, key input, asserted frames) | `dockerx/e2e_test.go`, `ui/e2e_test.go` |

Notes:
- E2E tests are behind the `//go:build e2e` tag, so `go test ./...` (the default,
  and the unit+functional CI job) never needs a daemon and never runs them.
- E2E fixtures are created/torn down with the `docker` CLI (`busybox`); the code
  under test is `dockerx` and the real `ui` program.
- The UI E2E uses `github.com/charmbracelet/x/exp/teatest` (Charm, MIT) — it runs
  the actual `tea.Program` in a simulated terminal, so it needs no TTY and is
  CI-stable. It is genuine end-to-end: real Docker client + real streaming.

Run `make test` (tiers 1–2) constantly; `make test-e2e` when you have a daemon.

## 5b. CI/CD (GitHub Actions)

Two workflows under `.github/workflows/`:

**`ci.yml`** — on push to `main` and on every PR:
- `lint` — `gofmt` check + `go vet` (default and `-tags e2e`).
- `test` — unit + functional on a matrix of **ubuntu / macOS / windows** (with
  `-race` except on Windows, where the race detector needs a C toolchain);
  uploads a coverage artifact.
- `e2e` — on ubuntu (Docker is preinstalled and running on GitHub-hosted Ubuntu);
  runs the `-tags e2e` suite against the real daemon.
- `release-dryrun` — runs GoReleaser in `--snapshot` mode and uploads the
  cross-platform archives as artifacts, so binary distribution is exercised on
  every change (not just on release).

**`release.yml`** — on pushing a `v*` tag:
- `verify` — gofmt + vet + the full test suite **including E2E**.
- `goreleaser` — builds for linux/darwin/windows × amd64/arm64, archives
  (`tar.gz`, `zip` on Windows) with `README`/`LICENSE`/`NOTICE`, a
  `checksums.txt`, and publishes a GitHub Release. Config: `.goreleaser.yaml`.

### Cutting a release

```sh
git tag v0.1.0
git push origin v0.1.0      # triggers release.yml → GitHub Release with binaries
```

Version/commit/date are injected into the binary via `-ldflags` and shown by
`contalyst --version`. Validate the release config locally any time with
`make snapshot` (or `goreleaser check`).

## 6. Conventions

- **Bubble Tea purity:** never block in `Update`; do I/O in a `tea.Cmd`. Mutate
  state by returning the updated model value.
- **Pointer-receiver helpers** (`setToast`, `applyTheme`, `rebuildList`,
  `recomputeLayout`, `teardownStreams`, `openConfirm`, `cycleTheme`) mutate the
  model in place; they're called on the addressable `m` inside value-receiver
  `Update`/handlers.
- **No SDK types in `ui`** — go through `dockerx` domain types.
- **Every destructive op** goes through `openConfirm`.
- **Streams** must be cancelable and torn down on view exit (`teardownStreams`).
- Run `make fmt vet test` before considering a change done.

---

## 7. Resuming cold

1. `make build && ./contalyst` to see current behavior; `make test` to confirm
   green.
2. Read [00-architecture.md](./00-architecture.md) §3–§5 (model, streams, list).
3. Pick the next backlog item in [03-known-issues.md](./03-known-issues.md).
4. Use the recipes above; keep the hint bar/help in sync with new keys.
