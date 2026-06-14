# Contalyst — Known Issues & Backlog

Current limitations, the status of every inception risk, and the prioritized
work that remains. Risk IDs (R1…R8) are defined in
[../inception/00-inception.md](../inception/00-inception.md) §8.

---

## 1. Known issues / limitations (current build)

| ID | Issue | Severity | Notes / mitigation |
|---|---|---|---|
| KI-1 | **Log-stream bleed on rapid detail switching.** Messages already buffered in a previous container's log channel can be appended to the new container's log view for a moment after switching. | Low | On switch, `teardownStreams` cancels the old context, but a few buffered `logLineMsg`s may already be in flight; `tea.Msg`s don't carry their channel identity. Fix: tag each stream with a generation counter on the model and drop messages from stale generations. |
| KI-2 | **`exec` requires the `docker` CLI** on `PATH` (it shells out via `tea.ExecProcess`). | Low | Intentional (CR-6). If absent, `e` produces an error toast. A pure-SDK `ContainerExecAttach` PTY path could remove the dependency. |
| KI-3 | **List refresh is poll-based** (1.5s `tea.Tick`), so state changes can lag up to ~1.5s and there's constant light polling. | Low | CR-7. Could subscribe to the Docker event stream for instant, cheaper updates. |
| KI-4 | ~~**`compactHints` has no key binding.**~~ **Resolved.** Bound to `H` (`keys.go` `CompactHints`, toggled in `update_list.go`). Frame decoration is now opt-in too, toggled with `F` (`keys.go` `Frame`, NFR-U5). | — | — |
| KI-5 | **Centered overlays replace the screen** rather than compositing over it (help/confirm). | Trivial (by design) | `overlayCenter` uses `lipgloss.Place`; true compositing needs a cell-merge overlay. Acceptable modal focus for now. |
| KI-6 | **No horizontal scroll / wrap in logs**; long lines are clipped to the viewport width. | Low | `viewport` shows lines as-is. Add soft-wrap or left/right scroll if needed. |
| KI-7 | **Images list uses `All:false`** (no intermediate/dangling images shown except via `prune`). | Trivial | Expose a toggle if intermediate images are wanted. |
| KI-8 | **[v2] Compose log aggregation (FR-CMP6) not wired.** Drilling into a project scopes the container list to its services; each service's logs open individually, but there is no single combined cross-service log stream. | Low | Add a project-level log view that multiplexes `docker compose logs -f` (or fan-in the per-service `LogStream`s) into one viewport. |
| KI-9 | **[v2] `ssh://` context switching needs an ssh connection helper** on `PATH` (docker's `connhelper`). Without it, `NewClientForHost` connects but the first call fails; surfaced as a toast, not a crash. | Low | Same class as KI-2 (CLI dependency). Document the requirement; unix/tcp endpoints are unaffected. |
| KI-10 | **[v2] Prune-dashboard reclaimable figures are best-effort.** They come from `/system/df`; where the daemon reports `-1`/unknown (RefCount, image Containers) the item is treated as in-use and excluded, so the displayed total can understate what `docker system prune` would free. | Low (by design) | Acceptable; the confirm dialog shows the estimate. Could call prune with `dry-run`-style accounting if the API gains it. |
| KI-11 | **[v2] Operation log is session-only** (in-memory ring buffer, cap 200; OQ-6). It is lost on quit. | Trivial (by design) | Log-to-file is out of scope (§5.3); revisit if persistence is wanted. |

None of these block use.

---

## 2. Inception risk status

How each inception risk (R1–R8) is handled in the build.

| Risk | Status | Where |
|---|---|---|
| R1 — log/stats demux wrong for TTY vs non-TTY | ✅ Handled | `dockerx/logs.go` branches on `HasTTY`; non-TTY via `stdcopy.StdCopy`. Verified by `TestSmoke`. |
| R2 — API-version mismatch crash on launch | ✅ Handled | `client.WithAPIVersionNegotiation()` in `dockerx/client.go`. |
| R3 — no logs/stats when running inside a container / over SSH | ⚠️ Designed for, not yet tested in-container | Uses env-based connection + negotiation; add an in-container test to fully close. |
| R4 — terminal-reserved key collisions | ✅ Handled | `keys.go` avoids `Ctrl-D/R/S`. (Rebindable keys = backlog.) |
| R5 — destructive default-button accident | ✅ Handled | `openConfirm` defaults the cursor to **Cancel**; covered by `TestConfirmDialogSafeDefault`. |
| R6 — two-line hint bar eats space | ✅ Handled | 1-line `compactHints` mode toggled with `H` (KI-4 resolved). |
| R7 — streaming goroutine leaks | ✅ Handled | Per-stream cancelable context; `teardownStreams` on view exit. |
| R8 — MIT vs Apache-2.0 license mismatch | ✅ Handled | Documented in DR-1; `LICENSE` (Apache-2.0) + `NOTICE` shipped. |
| R9 — compose shell-out depends on the `docker compose` plugin | ✅ Handled | `composeAvailCmd` probes once; `updateComposeActions` stays read-only and toasts when absent; the project list still groups from labels (CR-8/CR-9). |
| R10 — bulk partial failure | ✅ Handled | `bulkAction` runs per-target goroutines and reports `N ok / M failed` with the first error; never aborts mid-batch (CR-12). |
| R11 — context-switch stream leak / state bleed | ✅ Handled | `switchContext` calls `teardownStreams`, rebuilds the client, and resets kind/scope/marks/filter/data/cursor before reconnecting (CR-10). |
| R12 — destructive prune | ✅ Handled | Prune dashboard requires explicit per-category selection, shows estimated reclaimable space, and routes through the standard confirm dialog (safe default = Cancel, R5). |

---

## 3. Backlog (prioritized)

### v2 — shipped (M5–M8 / U9–U12)
The four areas promoted into scope by the 2026-06-14 inception update (DR-7) are
implemented and verified — see
[01-implementation-status.md](./01-implementation-status.md). Remaining
refinements within these areas are tracked above as KI-8…KI-11 (compose log
aggregation, ssh context helper, prune estimate accuracy, op-log persistence).

### Engineering backlog (not feature scope)
1. **Wire compose log aggregation** (KI-8) — the one remaining v2 sub-requirement (FR-CMP6).
2. **Fix KI-1** (stream generation tagging) — small, removes a correctness wart.
3. **Docker event-driven refresh** (KI-3) — replace/augment polling.
4. **In-container / SSH integration test** — closes R3 fully.

### Still out of scope (inception §5.3)
- Docker Swarm (nodes/services/stacks).
- Log save-to-file in detail view (in-log search already implemented — FR-C5, `/`+`n`/`N`).
- Full mouse support (row click, wheel scroll) — note it disables terminal text selection, so it would have to be opt-in.
- User-rebindable keybindings; externalize themes to a config file (inception OQ-3).

---

## 4. Open questions carried from inception

(see [../inception/00-inception.md](../inception/00-inception.md) §11)

- **OQ-1** Lip Gloss v1 vs v2 — **resolved**: v1 chosen (CR-1).
- **OQ-2** Confirm dialog: `huh` vs hand-rolled — **resolved**: hand-rolled
  overlay (`confirmBox`), no extra dependency.
- **OQ-3** Theme definition format — still embedded Go; externalization is out of scope.
- **OQ-4** Compose strategy: `docker compose` shell-out vs compose-go — **resolved**:
  shell-out chosen (CR-9 / DR-5); no new dependency. compose-go remains a future option.
- **OQ-5** Context-switch state: keep per-host vs reset — **resolved**: full reset
  on switch (CR-10 / R11), the simpler and leak-free choice.
- **OQ-6** Op-log persistence — **resolved**: session-only ring buffer (KI-11);
  file persistence is out of scope (§5.3).
