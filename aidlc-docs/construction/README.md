# Contalyst — Construction Artifacts

This directory is the AIDLC (AI-Driven Development Life Cycle) **Construction
phase** documentation — the **HOW**. The Inception phase (the WHAT / WHY) lives
in [`../inception/`](../inception/).

**Contalyst** is a modern, colorful TUI for managing Docker containers, written
in Go with the [Bubble Tea](https://github.com/charmbracelet/bubbletea)
framework. It lists containers/images/volumes/networks, streams logs and live
resource stats, runs lifecycle actions, and execs into a shell — keyboard-first,
discoverable, and usable over SSH.

These documents are written to be **self-contained and resumable**: a developer
(or AI agent) returning cold should be able to understand the architecture,
rebuild and test the project, and continue the work using only this directory
plus the inception docs.

## Documents (read in order)

| # | Document | Contents |
|---|---|---|
| 00 | [00-architecture.md](./00-architecture.md) | Package layout, the Elm-architecture data flow, streaming patterns, and construction decisions (CR-1…CR-12) |
| 01 | [01-implementation-status.md](./01-implementation-status.md) | Per–Unit-of-Work status mapped to files, requirement coverage, and test coverage |
| 02 | [02-developer-guide.md](./02-developer-guide.md) | Build / run / test, the dependency + license inventory, and step-by-step recipes for extending the app |
| 03 | [03-known-issues.md](./03-known-issues.md) | Known limitations, the status of every inception risk, and the prioritized backlog |

## Status at a glance

- **MVP + v2 complete** — Units of Work U0–U12 implemented and verified
  (see [01-implementation-status.md](./01-implementation-status.md)). v2 added
  docker-compose support, bulk/multi-select, multi-host/context switching, and
  the maintenance views (image layers, prune dashboard, operation log).
- **~3,850 lines** of Go (plus ~1,200 of tests) across `internal/dockerx`
  (Docker access) and `internal/ui` (Bubble Tea app).
- **Verified**: `go build`, `go vet` (incl. `-tags e2e`), `gofmt`, the unit +
  functional suite, and the E2E suite against a live Docker daemon all pass.

## Fastest way to resume

```sh
# from the repo root
make build && ./contalyst     # run it
make test                     # unit + functional tests (no daemon needed)
make test-e2e                 # end-to-end tests (needs a live daemon)
```

Then read [00-architecture.md](./00-architecture.md) and pick the next item from
the backlog in [03-known-issues.md](./03-known-issues.md).
