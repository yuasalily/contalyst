# Contalyst

[![CI](https://github.com/yuasalily/contalyst/actions/workflows/ci.yml/badge.svg)](https://github.com/yuasalily/contalyst/actions/workflows/ci.yml)
[![Release](https://github.com/yuasalily/contalyst/actions/workflows/release.yml/badge.svg)](https://github.com/yuasalily/contalyst/actions/workflows/release.yml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue)](./LICENSE)

A modern, colorful **Docker container management TUI** — built in Go with
[Bubble Tea](https://github.com/charmbracelet/bubbletea). Zellij-inspired: a
discoverable, keyboard-first interface with a colorful, always-visible hint bar.

Manage containers, stream logs, watch live resource stats, exec into a shell,
and browse images/volumes/networks — all without leaving the terminal, and all
over SSH.

```
Contalyst › Containers                                            docker 29.5.3 · Catalyst
   NAME            IMAGE            STATE     STATUS         PORTS
●  ct-web          nginx:alpine     running   Up 2 min       8088→80/tcp
○  ct-db           postgres:16      exited    Exited (0)
‖  ct-cache        redis:7          paused    Paused
↑↓ move · ⏎ logs · s start/stop · r restart · e exec · i inspect · d delete · / filter · : cmd
T theme · ? help · q quit
```

## Features

- **Live container list** with state shown by color *and* glyph (`●` running,
  `○` exited, `‖` paused, `◐` restarting/created).
- **Log streaming** with follow, scroll-to-pause, a timestamp toggle, and
  in-log search (`/`, then `n`/`N` to jump between matches) — correctly
  demultiplexed for both TTY and non-TTY containers.
- **Live stats** — CPU / memory bars, network and block I/O, PIDs.
- **Lifecycle actions** — start / stop / restart / pause / kill / remove, each
  destructive one guarded by a confirmation dialog that defaults to *Cancel*.
- **Exec** into a container shell (prefers `bash`, falls back to `sh`).
- **Inspect** any container as pretty-printed JSON.
- **Images, volumes, networks** — list, remove, and prune.
- **`/` fuzzy filter** and a **`:` command palette** (`:images`, `:volumes`,
  `:networks`, `:prune`, `:theme`, …) for fast navigation.
- **Themes** — Catalyst (default), Aurora, Mono — cycle live with `T`. Toggle a
  compact one-line hint bar with `H` and rounded/square frames with `F`.
- **Discoverable** — a context-sensitive hint bar is always on screen, and `?`
  opens full keybindings. No docs required.

## Install

A reachable Docker daemon is required at runtime.

**Prebuilt binary** (no Go needed) — download the archive for your OS/arch from
the [latest release](https://github.com/yuasalily/contalyst/releases/latest),
extract, and run. Verify against `checksums.txt`. Builds are published for
linux/macOS/Windows on amd64/arm64.

```sh
# example (linux amd64)
tar xzf contalyst_*_linux_amd64.tar.gz
./contalyst
```

**From source** (Go 1.26+):

```sh
go install github.com/yuasalily/contalyst@latest
# or, from a checkout:
make build && ./contalyst
```

Contalyst connects using the standard Docker environment (`DOCKER_HOST`, the
default socket, etc.) and negotiates the API version, so it works across a wide
range of Engine versions.

## Keybindings

| Key | Action | | Key | Action |
|---|---|---|---|---|
| `↑`/`k` `↓`/`j` | move | | `s` | start / stop |
| `g` / `G` | top / bottom | | `r` | restart |
| `⏎` / `l` | logs + stats | | `p` | pause / unpause |
| `i` | inspect | | `e` | exec shell |
| `/` | fuzzy filter | | `d` | remove |
| `:` | command palette | | `K` | kill |
| `T` | cycle theme | | `f` | toggle log follow |
| `R` | refresh | | `t` | toggle timestamps |
| `H` | compact hint bar | | `/` | search logs (in detail) |
| `F` | rounded/square frames | | `n` / `N` | next / prev log match |
| `?` | help | | `esc` | back / cancel |
| `q` | quit | | | |

## Architecture

```
main.go                 entry point
internal/dockerx/       Docker SDK isolated behind domain types (client, logs,
                        stats, resources) — the only package that imports the SDK
internal/ui/            Bubble Tea app (Elm architecture)
  app.go                root model, update routing, layout
  list.go / rows.go     custom colorful table + per-kind rows
  detail.go             logs + live stats split view
  update_list.go        list actions
  overlays.go           filter, command palette, confirm dialog, help, hint bar
  exec.go               exec via tea.ExecProcess
  theme/                color palettes + state→color/glyph mapping
```

Design rationale, requirements, and the work breakdown live in
[`aidlc-docs/inception/`](./aidlc-docs/inception/) (the WHAT/WHY); the
architecture, implementation status, developer guide, and known issues live in
[`aidlc-docs/construction/`](./aidlc-docs/construction/) (the HOW).

## Development

```sh
make test    # unit tests (no daemon needed)
make smoke   # integration test against the live daemon
make vet
```

## License

Apache-2.0. See [LICENSE](./LICENSE) and [NOTICE](./NOTICE). Third-party
dependencies are all permissive (Apache-2.0 / MIT / BSD); see NOTICE.
