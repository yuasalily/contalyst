# Contalyst — Implementation Status

Maps the inception **Units of Work** ([../inception/05-units-of-work.md](../inception/05-units-of-work.md))
to the code that implements them, with requirement coverage and test status.
Requirement IDs (FR-*/NFR-*) are defined in
[../inception/01-requirements.md](../inception/01-requirements.md).

**Overall: MVP complete; v2 (U9–U12) implemented and verified.** All units
U0–U12 are implemented. v2 added docker-compose support, bulk/multi-select,
multi-host/context switching, and the maintenance views (image layers, prune
dashboard, operation log) — see the v2 scope in
[../inception/00-inception.md](../inception/00-inception.md) §5.2.

---

## Unit status

| Unit | Title | Status | Primary files | Requirements covered |
|---|---|---|---|---|
| U0 | Foundation & Docker layer | ✅ Done | `main.go`, `dockerx/client.go`, `ui/app.go` | FR-E1/E2/E3, NFR-M1, NFR-M2, NFR-R1 |
| U1 | Container list & live updates | ✅ Done | `ui/list.go`, `ui/rows.go`, `dockerx/client.go` (`Containers`) | FR-C1, FR-C2, NFR-U2, NFR-P1 |
| U2 | Hint bar, help & theming | ✅ Done | `ui/overlays.go` (`hintView`/`helpBox`), `ui/theme/`, `ui/styles.go`, `ui/keys.go` (`CompactHints`/`Frame`) | FR-NAV5/6/7/9, NFR-U1/U3/U4/U5 |
| U3 | Log streaming & detail view | ✅ Done | `dockerx/logs.go`, `ui/detail.go` (`renderLogContent`/`computeMatches`/`highlightMatches`), `dockerx/client.go` (`Inspect`) | FR-C3, FR-C4, FR-C5, FR-C10, NFR-P2, NFR-R2 |
| U4 | Container controls & confirm dialogs | ✅ Done | `ui/update_list.go`, `ui/overlays.go` (confirm) | FR-C6, FR-C7, FR-C8, FR-NAV8, NFR-R3, NFR-R4 |
| U5 | Exec into shell | ✅ Done | `ui/exec.go` | FR-C9 |
| U6 | Stats streaming | ✅ Done | `dockerx/stats.go`, `ui/detail.go` (`statsContent`/`bar`) | FR-C11 |
| U7 | Filter & command palette | ✅ Done | `ui/overlays.go` (`updateFilter`/`updateCommand`), `ui/rows.go` (`fuzzyMatch`) | FR-NAV1/2/3/4, FR-NAV10 |
| U8 | Images / volumes / networks | ✅ Done | `dockerx/resources.go`, `ui/rows.go`, `ui/update_list.go` | FR-I1/I2, FR-V1, FR-N1 |
| U9 | docker-compose support | ✅ Done (FR-CMP6 partial) | `dockerx/compose.go`, `ui/rows.go` (`buildCompose`), `ui/update_list.go` (`updateComposeActions`) | FR-CMP1–5, FR-CMP7, NFR-CMP1, NFR-M5 — FR-CMP6 (cross-service log *aggregation*) deferred: drill-down scopes to the project's services and each opens its own log stream (KI-8) |
| U10 | Bulk / multi-select | ✅ Done | `ui/update_list.go` (`updateBulkActions`/`toggleMarkAll`), `ui/list.go` (mark gutter), `ui/messages.go` (`bulkAction`) | FR-B1–5, NFR-B1 |
| U11 | Multi-host / context switch | ✅ Done | `dockerx/context.go`, `ui/context.go`, `ui/app.go` (`reconnectedMsg`) | FR-H1–5, NFR-H1 |
| U12 | Maintenance: layers / prune / op log | ✅ Done | `dockerx/maintenance.go`, `ui/maintenance.go`, `ui/app.go` (`recordOp`) | FR-L1, FR-PR1/2, FR-OL1/2 |

---

## Feature checklist (mapped to inception scope §5.1)

- [x] Container list (live, auto-refresh every 1.5s) with state color + glyph
- [x] Log streaming: follow, scroll-to-pause, timestamp toggle, in-log search (`/` + `n`/`N`), TTY/non-TTY demux
- [x] Live stats: CPU/mem bars, net, block I/O, PIDs
- [x] Lifecycle: start / stop / restart / pause / unpause / kill / remove
- [x] Inspect (pretty JSON, scrollable)
- [x] Exec into shell (bash→sh fallback)
- [x] Images / volumes / networks: list + remove + prune
- [x] `/` fuzzy filter (live), `:` command palette (with tab-autosuggest)
- [x] Persistent hint bar + `?` full help
- [x] Confirmation dialogs on destructive actions, default focus on **Cancel**
- [x] Themes (Catalyst/Aurora/Mono), live switch with `T`
- [x] Compact 1-line hint bar toggle (`H`) and rounded/square frame toggle (`F`)
- [x] Connection-failure screen (no crash)

### v2 (inception scope §5.2)

- [x] **Compose** (`:compose`): project list grouped by label, state color (up/degraded/down), drill-down to scoped services, `u` up / `d` down (confirmed) / `r` restart / `b` build / `B` build --no-cache, read-only fallback when the plugin is missing
- [x] **Bulk/multi-select**: `space` mark, `a` mark-all/clear, `esc` clear, bulk `s`/`S`/`r`/`d` on the marked set, count-aware confirm on remove, concurrent fan-out with `N ok / M failed` aggregation
- [x] **Multi-host**: `:context`/`:hosts` switcher overlay, active host in the header, runtime client swap with stream teardown + state reset
- [x] **Maintenance**: image-layer view (`⏎` on an image), prune dashboard (`:prune`, per-category reclaimable + select + confirm), operation log (`@`, session ring buffer)

---

## Test coverage

Three tiers (full strategy in [02-developer-guide.md](./02-developer-guide.md)
§5a). Unit + functional need no daemon; E2E is behind the `e2e` build tag.

**Unit** (`-run TestUnit`, no daemon) — pure functions:

| File | Covers |
|---|---|
| `dockerx/unit_test.go` | `primaryName`, `shortImage`, `formatPorts`, `computeStats` (CPU%/mem/net/blk math, no-delta case) |
| `dockerx/compose_unit_test.go` | **[v2]** `ComposeProjects` grouping/state, `composeState`, `composeBaseArgs`, `ComposeOp.args`, `lastLine` |
| `ui/unit_test.go` | `pad`, `resolveWidths`, `fuzzyMatch`, `commandSuggest`, `humanSize`, `relativeTime`, `packHints`, `highlightMatches` |
| `ui/v2_unit_test.go` | **[v2]** `formatLayers`, `plural`, `truncMiddle`, `composeStateGlyph` |
| `ui/theme/theme_test.go` | `StateColor`, `StateGlyph`, `Next` cycling |

**Functional** (`-run TestFunctional`, no daemon) — `ui/functional_test.go` and
`ui/v2_functional_test.go` drive the model through `Update`/`View` with synthetic
messages (`feed()` helper, `nil` client): list render, detail (logs+stats), help,
**confirm-dialog safe default**, theme cycle, filter, command-palette switch,
in-log search (`/`+`n`/`N`), compact-hint toggle, frame-style toggle,
tiny-terminal no-panic. **v2**: compose project list + drill-down scope, bulk
mark + count-aware delete confirm, mark-all/clear, context switcher overlay +
header host, prune dashboard select→confirm, operation log capture, image-layer
view.

**E2E** (`-tags e2e -run TestE2E`, **needs Docker**):

| File / test | Covers |
|---|---|
| `dockerx/e2e_test.go` · `TestE2E_ContainerLifecycleAndStreams` | Create real container → list → **log demux** → stats → inspect → stop → remove |
| `dockerx/e2e_test.go` · `TestE2E_ResourceLists` | Images/volumes/networks list against live daemon |
| `dockerx/e2e_test.go` · `TestE2E_MaintenanceReads` | **[v2]** `ImageHistory` + `DiskUsage` (5 prune categories) against live daemon |
| `dockerx/e2e_test.go` · `TestE2E_ComposeAndContexts` | **[v2]** `ComposeAvailable` probe + `Contexts` enumeration degrade gracefully |
| `ui/e2e_test.go` · `TestE2E_TUIShowsContainerAndOpensLogs` | Runs the real program via `teatest`: shows the container, filters, opens logs, sees streaming output, quits |

Run: `make test` (unit+functional), `make test-e2e` (E2E), `make cover`,
`make lint`. All three tiers also run in CI (`.github/workflows/ci.yml`).

---

## Not yet implemented

**v2 (M5–M8 / U9–U12) is now implemented** (see the unit table above). A few
refinements are tracked as known issues in
[03-known-issues.md](./03-known-issues.md) (KI-8…KI-11): compose project log
*aggregation* in the detail view is not yet wired (only per-container logs),
the prune-dashboard reclaimable figures are best-effort estimates, the op log is
session-only, and ssh:// context switching depends on an ssh connection helper.

**Still out of scope** (inception §5.3):

- Docker Swarm (nodes/services/stacks)
- Save logs to file
- Full mouse support; user-rebindable keys; config-file customization
