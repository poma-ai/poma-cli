# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build -o poma .        # Build the binary (Go 1.21+)
go build ./...            # Build all packages
go vet ./...              # Static analysis
go test ./...             # Run all tests
go test ./pkg/client/...  # Run a specific package's tests
```

Release is handled by GoReleaser on tagged commits via GitHub Actions — don't run it manually.

## Architecture

**poma-cli** is a Go CLI (Cobra) for the POMA AI document ingestion API. It has two main packages:

- **`internal/cli/`** — Cobra command tree. `root.go` sets up persistent flags (`--base-url`, `--token`, `--json`) and a `PersistentPreRunE` hook that merges `--json` config into flags. Commands are grouped in `account.go`, `job.go`, and `health.go`. All commands use `RunE` and return errors; never call `os.Exit` in command handlers.

- **`pkg/client/`** — HTTP client wrapping `net/http`. `client.go` has one method per API endpoint. `models.go` holds request/response structs. `safety.go` contains all input validators and `FileConfig` (the struct backing `--json` flag). `pathseg.go` handles job ID URL normalization.

**`--json` flag flow:** User passes inline JSON or a `.json` file path → `config.go:mergeConfigIntoFlags` overlays values onto explicit flags (explicit flags win) → `PersistentPreRunE` validates merged values via `client.ValidateFileConfig`.

## Extending the code

**New API endpoint:**
1. Add request/response structs in `pkg/client/models.go`
2. Add method on `Client` in `pkg/client/client.go` using `Do`/`DoJSON`
3. Use `JobPathSegment` for job ID path segments (not raw `EncodePathSegment`)
4. Add Cobra command in the appropriate `internal/cli/*.go` file, registered in the parent command's constructor
5. Validate all inputs before calling the client (`ValidateJobID`, `ValidateSafeOutputDir`, `ValidateUserStrings`)

**New `--json` flag field:**
1. Add field to `FileConfig` in `pkg/client/safety.go`
2. Wire it in `mergeConfigIntoFlags` in `internal/cli/config.go`
3. Add validation in `client.ValidateFileConfig`

## Security rules (non-negotiable)

- All flag and `--json` values are untrusted — validate before use
- Never log or echo full JWTs or `api_key` values
- Output paths must stay under CWD (existing helpers enforce this; maintain the pattern)

## Documentation

When adding or changing a command or flag, update:
- `README.md` for user-facing changes
- `AGENTS.md` for command table, flags, and agent-oriented behavior notes
