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

None of these block MVP use.

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

---

## 3. Backlog (prioritized)

Derived from inception scope §5.2 and the issues above.

### P1 — highest value
1. **docker-compose first-class support** — up/down/restart/rebuild (incl.
   `--no-cache`), dependency-aware. The biggest gap competitors leave open
   (inception research). New `dockerx` calls + a compose resource kind.
2. **Fix KI-1** (stream generation tagging) — small, removes a correctness wart.

### P2
3. **Docker event-driven refresh** (KI-3) — replace/augment polling.
4. **Bulk / multi-select actions** — multi-select in `list.go`, batch via `parallel`-style command fan-out.
5. **Log save-to-file** in detail view (in-log search already implemented — FR-C5, `/`+`n`/`N`).
6. **In-container / SSH integration test** — closes R3 fully.

### P3
8. Multiple hosts / context switching.
9. Image-layer view; prune dashboards; operation/command log (lazygit `@`).
10. Full mouse support (row click, wheel scroll) — note it disables terminal text selection, so make it opt-in.
11. User-rebindable keybindings (config file).
12. Externalize themes to a config file (inception OQ-3).

---

## 4. Open questions carried from inception

(see [../inception/00-inception.md](../inception/00-inception.md) §11)

- **OQ-1** Lip Gloss v1 vs v2 — **resolved**: v1 chosen (CR-1).
- **OQ-2** Confirm dialog: `huh` vs hand-rolled — **resolved**: hand-rolled
  overlay (`confirmBox`), no extra dependency.
- **OQ-3** Theme definition format — still embedded Go; externalization is P3.
