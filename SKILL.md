# Skills (poma-cli)

Short checklist for humans and **AI agents** working on this repository. For full API and command semantics, use **[AGENTS.md](./AGENTS.md)** and **[README.md](./README.md)**.

---

## Operating the CLI

- **Build:** `go build -o poma .` from repo root (Go 1.21+, see `go.mod`).
- **Auth:** After `account verify-email`, generate a long-lived opaque API key with **`account generate-api-key`** (`POST /generateApiKey`). Export **`POMA_API_KEY`** or pass **`--token`** / **`--json`**.
- **Defaults:** `--base-url` and `--status-base-url` point at production; override for staging or mocks.
- **Machine-friendly output:** Commands print pretty JSON to stdout; **`account generate-api-key`** emits only `{"api_key":"…"}`.

---

## Extending the HTTP client (`pkg/client`)

1. Add request/response structs in **`pkg/client/models.go`** when JSON shapes matter.
2. Add a method on **`Client`** in **`client.go`** (or a small focused file) using **`Do`** / **`DoJSON`**.
3. For **`job_id`** and other UUID path segments, use **`JobPathSegment`** (normalizes API quirks + safe for strict UUID paths).
4. Ingest uses **`sanitizeContentDispositionFilename`** for the uploaded basename; keep that pattern for new upload-style endpoints.
5. For new input file paths, use **`ValidateInputFilePath`**; for output paths use **`ValidateSafeOutputDir`**.

---

## Extending the CLI (`internal/cli`)

1. Add **`cobra.Command`** in the right file (`account.go`, `primecut.go`, `job.go`, `orga.go`, `project.go`, `cheatsheet.go`, …) or a new file if the group is large.
2. Register the command on the parent in the parent's constructor (e.g. **`AccountCmd()`** → **`cmd.AddCommand(...)`**).
3. Reuse **`apiClient()`** and **`PrintJSON`**; return errors from **`RunE`** (don't **`os.Exit`** in commands).
4. **Validate inputs** before calling the client:
   - **`client.ValidateJobID`** for job IDs
   - **`client.ValidateResourceName`** for UUID/slug path parameters (orga IDs, project IDs, account IDs)
   - **`client.ValidateSafeOutputDir`** for any path written under user control (downloads)
   - **`client.ValidateInputFilePath`** for any file path read under user control (`--input`)
   - **`client.ValidateUserStrings`** for email and free-text flags
   - New flags used from **`--json`**: add fields to **`client.FileConfig`** in **`pkg/client/safety.go`**, wire **`mergeConfigIntoFlags`** in **`config.go`**, and **`client.ValidateFileConfig`**.
5. Run **`go build ./...`** and **`go vet ./...`**.

---

## Security posture (non-negotiable)

- Treat all flag and **`--json`** values as **untrusted** (agents may hallucinate paths, IDs, or encodings).
- Do not log or echo full JWTs / **`api_key`** in new code paths.
- **Output paths** and **`--json` file paths** must stay under **CWD** (existing helpers enforce this).
- **`--input` file paths** (cheatsheet) must also stay under **CWD** (`ValidateInputFilePath`).

---

## Documentation touchpoints

| Change | Update |
|--------|--------|
| New or changed command / flag | **README.md** (if user-facing), **AGENTS.md** (command table + flags + behavior notes) |
| Auth or token flow | **README.md**, **AGENTS.md** |
| New file added to `internal/cli/` or `pkg/client/` | **README.md** project structure, **SKILL.md** file map |
| Agent-oriented guardrails | **AGENTS.md**, this file |

---

## Quick file map

| Path | Role |
|------|------|
| `main.go` | Entrypoint |
| `internal/cli/root.go` | Root command, persistent flags, **`--json`** merge hook |
| `internal/cli/account.go` | Registration, verify, me, generate-api-key, my-projects, my-usage |
| `internal/cli/primecut.go` | PrimeCut **`ingest`** / **`ingest-sync`** commands |
| `internal/cli/job.go` | Job lifecycle: status, status-stream, result, download, delete |
| `internal/cli/orga.go` | Organisation CRUD, members, invitations, accept-invitation |
| `internal/cli/project.go` | Project CRUD and search |
| `internal/cli/cheatsheet.go` | Cheatsheet create (local, no API call) |
| `internal/cli/health.go` | Health check |
| `internal/cli/config.go` | JSON config shape and merge into flags |
| `internal/cli/util.go` | Shared helpers (`PrintJSON`, `requireToken`) |
| `pkg/client/client.go` | HTTP transport and all endpoint methods |
| `pkg/client/models.go` | Request/response structs |
| `pkg/client/cheatsheet.go` | Cheatsheet generation algorithm (port of `bin/sdk/retrieval.py`) |
| `pkg/client/cheatsheet_test.go` | Unit tests for cheatsheet generation |
| `pkg/client/pathseg.go` | URL path segment encoding |
| `pkg/client/safety.go` | Input validation, `FileConfig` for `--json` |
