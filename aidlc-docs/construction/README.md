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
| 00 | [00-architecture.md](./00-architecture.md) | Package layout, the Elm-architecture data flow, streaming patterns, and construction decisions (CR-1…CR-7) |
| 01 | [01-implementation-status.md](./01-implementation-status.md) | Per–Unit-of-Work status mapped to files, requirement coverage, and test coverage |
| 02 | [02-developer-guide.md](./02-developer-guide.md) | Build / run / test, the dependency + license inventory, and step-by-step recipes for extending the app |
| 03 | [03-known-issues.md](./03-known-issues.md) | Known limitations, the status of every inception risk, and the prioritized backlog |

## Status at a glance

- **MVP complete** — all Units of Work U0–U8 implemented and verified
  (see [01-implementation-status.md](./01-implementation-status.md)).
- **~2,400 lines** of Go across `internal/dockerx` (Docker access) and
  `internal/ui` (Bubble Tea app).
- **Verified**: `go build`, `go vet`, `gofmt`, and the test suite all pass; an
  integration smoke test runs against a live Docker daemon.

## Fastest way to resume

```sh
# from the repo root
make build && ./contalyst     # run it
make test                     # unit tests (no daemon needed)
make smoke                    # integration test (needs a live daemon)
```

Then read [00-architecture.md](./00-architecture.md) and pick the next item from
the backlog in [03-known-issues.md](./03-known-issues.md).
