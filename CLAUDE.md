# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
# Build the binary
make build          # outputs ./dev
make devbin         # outputs ./devbin/dev (used for self-installation)

# Run tests
go test ./...

# Run a single test
go test ./pkg/executor/ -run TestNormalizeClonePath

# Run tests with verbose output
go test -v ./...
```

## Architecture

`dev` is a CLI tool that organizes cloned git repositories into a directory tree mirroring the repo URL, inspired by the old `go get` behavior. It requires a `$DEVPATH` environment variable pointing to the base directory.

**Entry point:** `main.go` validates `$DEVPATH`, ensures `src/` and `bin/` subdirectories exist, then delegates to `pkg/executor`.

**`pkg/executor/executor.go`** contains all command logic:

- `Executor` struct holds path config (`devPath`, `workDir = $DEVPATH/src`, `binDir`, `devbinDir`) and the resolved command.
- Three commands: `get` (clone), `install` (clone + `make devbin` + copy binaries to `$DEVPATH/bin`), `init` (create directory + `git init` + optional `go mod init` + initial commit).
- `normalizeClonePath` handles three input formats: plain `host/org/repo`, `https://...`, and `git@host:org/repo` SSH URLs — all normalized to a consistent `host/org/repo` directory path under `$DEVPATH/src`.
- All git/make commands are run via `runCommand`, which sets `cmd.Dir` to the absolute target path (no `os.Chdir` calls anywhere).

**Install flow:** `install` runs `make devbin` in the cloned repo, then walks the `devbin/` subdirectory and moves all files to `$DEVPATH/bin`.
