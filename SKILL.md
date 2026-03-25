# Skills (poma-cli)

Short checklist for humans and **AI agents** working on this repository. For full API and command semantics, use **[AGENTS.md](./AGENTS.md)** and **[README.md](./README.md)**.

---

## Operating the CLI

- **Build:** `go build -o poma .` from repo root (Go 1.21+, see `go.mod`).
- **Auth:** After `account verify-email`, prefer long-lived **`api_key`** via **`account api-key`** (`GET /me`). Export **`POMA_API_KEY`** or pass **`--token`** / **`--json`**.
- **Defaults:** `--base-url` and `--status-base-url` point at production; override for staging or mocks.
- **Machine-friendly output:** Commands print pretty JSON to stdout; **`account api-key`** emits only `{"api_key":"…"}`.

---

## Extending the HTTP client (`pkg/client`)

1. Add request/response structs in **`pkg/client/models.go`** when JSON shapes matter.
2. Add a method on **`Client`** in **`client.go`** (or a small focused file) using **`Do`** / **`DoJSON`**.
3. For **`job_id`** path segments, use **`JobPathSegment`** (normalizes API quirks + safe for strict UUID paths). **`EncodePathSegment`** remains for other escaped segments.
4. Ingest uses **`sanitizeContentDispositionFilename`** for the uploaded basename; keep that pattern for new upload-style endpoints.

---

## Extending the CLI (`internal/cli`)

1. Add **`cobra.Command`** in the right file (`account.go`, `jobs.go`, …) or a new file if the group is large.
2. Register the command on the parent in the parent’s constructor (e.g. **`AccountCmd()`** → **`cmd.AddCommand(...)`**).
3. Reuse **`apiClient()`** and **`PrintJSON`**; return errors from **`RunE`** (don’t **`os.Exit`** in commands).
4. **Validate inputs** before calling the client:
   - **`client.ValidateJobID`** for job IDs
   - **`client.ValidateSafeOutputDir`** for any path written under user control (downloads)
   - **`client.ValidateUserStrings`** (and related) for text flags as appropriate
   - New flags used from **`--json`**: add fields to **`client.FileConfig`** in **`pkg/client/safety.go`**, wire **`mergeConfigIntoFlags`** in **`config.go`**, and **`client.ValidateFileConfig`**.
5. Run **`go build ./...`** and **`go vet ./...`**.

---

## Security posture (non-negotiable)

- Treat all flag and **`--json`** values as **untrusted** (agents may hallucinate paths, IDs, or encodings).
- Do not log or echo full JWTs / **`api_key`** in new code paths.
- **Output paths** and **`--json` file paths** must stay under **CWD** (existing helpers enforce this).

---

## Documentation touchpoints

| Change | Update |
|--------|--------|
| New or changed command / flag | **README.md** (if user-facing), **AGENTS.md** (command table + behavior) |
| Auth or token flow | **README.md**, **AGENTS.md** |
| Agent-oriented guardrails | **AGENTS.md**, this file |

---

## Quick file map

| Path | Role |
|------|------|
| `main.go` | Entrypoint |
| `internal/cli/root.go` | Root command, persistent flags, **`--json`** merge hook |
| `pkg/client/safety.go` | Input validation, `FileConfig` for `--json` |
| `internal/cli/config.go` | JSON config shape and merge into flags |
| `pkg/client/client.go` | HTTP transport and endpoint methods |
| `pkg/client/pathseg.go` | URL path segment encoding |
