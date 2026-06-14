# Contalyst — Implementation Status

Maps the inception **Units of Work** ([../inception/05-units-of-work.md](../inception/05-units-of-work.md))
to the code that implements them, with requirement coverage and test status.
Requirement IDs (FR-*/NFR-*) are defined in
[../inception/01-requirements.md](../inception/01-requirements.md).

**Overall: MVP complete.** All units U0–U8 are implemented and verified.

---

## Unit status

| Unit | Title | Status | Primary files | Requirements covered |
|---|---|---|---|---|
| U0 | Foundation & Docker layer | ✅ Done | `main.go`, `dockerx/client.go`, `ui/app.go` | FR-E1/E2/E3, NFR-M1, NFR-M2, NFR-R1 |
| U1 | Container list & live updates | ✅ Done | `ui/list.go`, `ui/rows.go`, `dockerx/client.go` (`Containers`) | FR-C1, FR-C2, NFR-U2, NFR-P1 |
| U2 | Hint bar, help & theming | ✅ Done | `ui/overlays.go` (`hintView`/`helpBox`), `ui/theme/`, `ui/styles.go` | FR-NAV5/6/7/9, NFR-U1/U3/U4/U5 |
| U3 | Log streaming & detail view | ✅ Done | `dockerx/logs.go`, `ui/detail.go`, `dockerx/client.go` (`Inspect`) | FR-C3, FR-C4, FR-C5, FR-C10, NFR-P2, NFR-R2 |
| U4 | Container controls & confirm dialogs | ✅ Done | `ui/update_list.go`, `ui/overlays.go` (confirm) | FR-C6, FR-C7, FR-C8, FR-NAV8, NFR-R3, NFR-R4 |
| U5 | Exec into shell | ✅ Done | `ui/exec.go` | FR-C9 |
| U6 | Stats streaming | ✅ Done | `dockerx/stats.go`, `ui/detail.go` (`statsContent`/`bar`) | FR-C11 |
| U7 | Filter & command palette | ✅ Done | `ui/overlays.go` (`updateFilter`/`updateCommand`), `ui/rows.go` (`fuzzyMatch`) | FR-NAV1/2/3/4, FR-NAV10 |
| U8 | Images / volumes / networks | ✅ Done | `dockerx/resources.go`, `ui/rows.go`, `ui/update_list.go` | FR-I1/I2, FR-V1, FR-N1 |

---

## Feature checklist (mapped to inception scope §5.1)

- [x] Container list (live, auto-refresh every 1.5s) with state color + glyph
- [x] Log streaming: follow, scroll-to-pause, timestamp toggle, TTY/non-TTY demux
- [x] Live stats: CPU/mem bars, net, block I/O, PIDs
- [x] Lifecycle: start / stop / restart / pause / unpause / kill / remove
- [x] Inspect (pretty JSON, scrollable)
- [x] Exec into shell (bash→sh fallback)
- [x] Images / volumes / networks: list + remove + prune
- [x] `/` fuzzy filter (live), `:` command palette (with tab-autosuggest)
- [x] Persistent hint bar + `?` full help
- [x] Confirmation dialogs on destructive actions, default focus on **Cancel**
- [x] Themes (Catalyst/Aurora/Mono), live switch with `T`
- [x] Connection-failure screen (no crash)

---

## Test coverage

Three tiers (full strategy in [02-developer-guide.md](./02-developer-guide.md)
§5a). Unit + functional need no daemon; E2E is behind the `e2e` build tag.

**Unit** (`-run TestUnit`, no daemon) — pure functions:

| File | Covers |
|---|---|
| `dockerx/unit_test.go` | `primaryName`, `shortImage`, `formatPorts`, `computeStats` (CPU%/mem/net/blk math, no-delta case) |
| `ui/unit_test.go` | `pad`, `resolveWidths`, `fuzzyMatch`, `commandSuggest`, `humanSize`, `relativeTime`, `packHints` |
| `ui/theme/theme_test.go` | `StateColor`, `StateGlyph`, `Next` cycling |

**Functional** (`-run TestFunctional`, no daemon) — `ui/functional_test.go` drives
the model through `Update`/`View` with synthetic messages (`feed()` helper, `nil`
client): list render, detail (logs+stats), help, **confirm-dialog safe default**,
theme cycle, filter, command-palette switch, tiny-terminal no-panic.

**E2E** (`-tags e2e -run TestE2E`, **needs Docker**):

| File / test | Covers |
|---|---|
| `dockerx/e2e_test.go` · `TestE2E_ContainerLifecycleAndStreams` | Create real container → list → **log demux** → stats → inspect → stop → remove |
| `dockerx/e2e_test.go` · `TestE2E_ResourceLists` | Images/volumes/networks list against live daemon |
| `ui/e2e_test.go` · `TestE2E_TUIShowsContainerAndOpensLogs` | Runs the real program via `teatest`: shows the container, filters, opens logs, sees streaming output, quits |

Run: `make test` (unit+functional), `make test-e2e` (E2E), `make cover`,
`make lint`. All three tiers also run in CI (`.github/workflows/ci.yml`).

---

## Not yet implemented (deferred to Post-MVP)

Per inception scope §5.2; tracked with priorities in
[03-known-issues.md](./03-known-issues.md):

- docker-compose first-class support (up/down/rebuild --no-cache, dependency order)
- Docker Swarm (nodes/services/stacks)
- Multiple hosts / context switching
- Bulk / multi-select actions
- Image-layer view, prune dashboards, operation/command log
- Save logs to file
- Full mouse support; user-rebindable keys; compact-hint-bar toggle binding
  (`compactHints` exists on the model but has no key bound yet)
