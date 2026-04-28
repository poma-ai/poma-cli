# Agents

This document describes the **implemented** POMA CLI (`poma`): responsibilities, API usage, and constraints. The CLI talks to the [Poma AI REST API](https://api.poma-ai.com/v3) for ingestion, job lifecycle, account management, organisations, projects, and cheatsheet generation.

**This CLI is frequently invoked by AI/LLM agents. Always assume inputs can be adversarial** — see [Input safety](#input-safety) and `pkg/client/safety.go`.

---

## Project overview

Go + Cobra binary that wraps the public API: register/verify email, authenticated account endpoints, PrimeCut ingest (pro/eco), job status (one-shot or SSE), result, download, delete, organisation management, project management, local cheatsheet generation, and health. Responses are pretty-printed JSON on stdout unless noted.

**Layout**

- `main.go` — entrypoint, calls `internal/cli.Execute()`
- `internal/cli/` — Cobra command tree (`root`, `account`, `primecut`, `job`, `orga`, `project`, `cheatsheet`, `health`), `--json` merge (`config.go`)
- `pkg/client/` — HTTP client (`net/http`), models, path-segment encoding (`pathseg.go`), input validation (`safety.go`, `FileConfig`), cheatsheet generation (`cheatsheet.go`)

---

## Stack & constraints

| Area | Implementation |
|------|----------------|
| **Language** | Go 1.21+ (`go.mod`) |
| **CLI** | [Cobra](https://github.com/spf13/cobra) |
| **HTTP** | `net/http`; `Authorization: Bearer <token>` when `Client.Token` is set |
| **Config / secrets** | Flags, `POMA_API_KEY`, and optional `--json` (see below). **No** on-disk config file in this repo. |
| **Output** | `fmt` + `PrintJSON` for API bodies; errors to stderr via caller (`main` prints error and exits 1) |

---

## Global flags (all subcommands)

| Flag | Default | Notes |
|------|---------|--------|
| `--base-url` | `https://api.poma-ai.com/v3` | Must be `http` or `https` with a host. |
| `--status-base-url` | `https://api.poma-ai.com/status/v1` | Used by `job status-stream` and `primecut ingest-sync`. |
| `--token` | `$POMA_API_KEY` | Required for authenticated routes. |
| `--json` | (empty) | Inline JSON object (`{...}`) **or** path to a `.json` file under the process **CWD**; keys map to flags; explicit flags **override** JSON. |

### `--json` keys (snake_case)

`base_url`, `status_base_url`, `token`, `email`, `username`, `company`, `code`, `file`, `job_id`, `output` — only flags that exist on the invoked command are applied.

---

## Authentication flow

1. `poma account register-email --email …` → `POST /registerEmail` (optional `--username`, `--company`). No JWT.
2. `poma account verify-email --email … --code …` → `POST /verifyEmail`; response includes a **token** (JWT). The command prints `Token: …` and the JSON body. Use this as a **bootstrap** credential.
3. With that bearer token, generate a long-lived opaque API key: **`poma account generate-api-key`** → `POST /generateApiKey`; prints `{"api_key":"…"}`. This key is suitable for ongoing CLI use and automation. Set it as `POMA_API_KEY`.
4. The CLI does **not** persist tokens to a file; callers supply flags, `POMA_API_KEY`, or `--json`.
5. Authenticated calls use `Authorization: Bearer <token>` (same header for the bootstrap JWT or the opaque key).

**Automation hint:** `export POMA_API_KEY=$(poma account generate-api-key | jq -r '.api_key')`

---

## Input safety

Implemented in `pkg/client/safety.go` and `pkg/client` (path segments, `Content-Disposition` filename):

- **C0 controls** (ASCII `0x00`–`0x1F`): rejected on most flag/JSON strings; inline `--json` allows tab/LF/CR only.
- **Job IDs** (`--job-id`, JSON `job_id`): no `?` `#` `%`, no path separators; must be a single path segment for URLs.
- **Download output** (`--output` / JSON `output`): resolved and **must stay under the process current working directory** (same rule for `--json` when it is a **file path**).
- **`--input` file paths** (cheatsheet): resolved and must stay under CWD (`ValidateInputFilePath`).
- **HTTP paths**: job IDs and resource IDs are passed through `url.PathEscape` in `pkg/client` when building URLs.

Ingest `--file` rejects control characters but **may** be any readable path (not restricted to CWD).

---

## Command reference (by area)

Below, "agent" names are **logical groupings** for automation docs; each maps to real `poma …` subcommands.

### `account` — registration, verification, and account data

| Command | API | Auth |
|---------|-----|------|
| `poma account register-email` | `POST /registerEmail` | No |
| `poma account verify-email` | `POST /verifyEmail` | No |
| `poma account me` | `GET /me` | JWT |
| `poma account generate-api-key` | `POST /generateApiKey` | JWT |
| `poma account my-projects` | `GET /myProjects` | JWT |
| `poma account my-usage` | `GET /myUsage` | JWT |

**Flags**

- `register-email`: `--email` / `-e` (required), `--username` / `-u`, `--company` / `-c`
- `verify-email`: `--email` / `-e`, `--code` / `-k` (both required)
- `me`, `generate-api-key`, `my-projects`, `my-usage`: no subcommand-specific flags beyond globals. Missing token → error.

**Verify output:** prints `Token:` line plus JSON (includes JWT). Use this next to call `poma account generate-api-key` to get a long-lived key. **Do not** log either value in shared transcripts.

**`generate-api-key`:** calls `POST /generateApiKey` and prints `{"api_key":"…"}`. This rotates (or creates) the account's opaque API key. Use this value as `POMA_API_KEY` / `--token` for new shells and automation. **Do not** log or commit the key.

---

### `primecut` — ingest and ingest-sync

| Command | API | Auth |
|---------|-----|------|
| `poma primecut ingest` | `POST /primeCut/ingest` or `POST /primeCutEco/ingest` (**`--eco`** or aliases **`ingest-eco`** / **`ingest-eco-data`**). Aliases: **`ingest-data`**, **`ingest-eco`**, **`ingest-eco-data`** | JWT |
| `poma primecut ingest-sync` | `POST /primeCut/ingest` or `POST /primeCutEco/ingest` (**`--eco`**), then SSE until terminal; if `done` and **`--output`** set → `GET /jobs/{job_id}/download`; if `done` and no **`--output`** → `GET /jobs/{job_id}/results` (JSON to stdout) | JWT |

**Flags**

- `ingest`: either **`--file` / `-f`** (path to file), or **`--filename` / `-n`** plus **`--data`** or stdin; **`--eco`** for eco endpoint
- `ingest-sync`: same input modes; **`--eco`**; **`--output` / `-o`** (optional; when set, downloads the archive to that path; when omitted, fetches results JSON)

**Behavior notes**

- **Ingest:** body held in memory; `Content-Disposition` from file basename or `--filename`. Prints only `{"job_id":"…"}`.
- **ingest-sync:** ingest → SSE stream until terminal status. `done` + `--output` → download archive; `done` without `--output` → fetch and print results JSON. `failed`/`deleted` → error.

---

### `job` — status, result, download, delete

| Command | API | Auth |
|---------|-----|------|
| `poma job status` | `GET /jobs/{job_id}/status` | JWT |
| `poma job status-stream` | SSE `GET {status-base-url}/jobs/{job_id}` | JWT |
| `poma job result` | `GET /jobs/{job_id}/results` | JWT |
| `poma job download` | `GET /jobs/{job_id}/download` | JWT |
| `poma job delete` | `DELETE /jobs/{job_id}` | JWT |

**Flags**

- `status`, `status-stream`, `result`, `delete`: `--job-id` (required)
- `download`: `--job-id` (required), `--output` / `-o` (optional; default `bin/{job_id}.poma` under CWD)

**Behavior notes**

- **Status:** single request; no polling. Wrap in a shell loop if needed.
- **Status-stream:** reads SSE until terminal `job_status` (`done`, `failed`, `deleted`) or EOF; each event printed as JSON.
- **Download:** response fully buffered then written to path; no `--force` (overwrites silently).
- **Delete:** prints short confirmation on HTTP 200.

---

### `orga` — organisation management

| Command | API | Auth |
|---------|-----|------|
| `poma orga list` | `GET /orgas` | JWT |
| `poma orga create` | `POST /orgas` | JWT |
| `poma orga get` | `GET /orgas/{orgaId}` | JWT |
| `poma orga update` | `PUT /orgas/{orgaId}` | JWT |
| `poma orga delete` | `DELETE /orgas/{orgaId}` | JWT |
| `poma orga members list` | `GET /orgas/{orgaId}/members` | JWT |
| `poma orga members add` | `POST /orgas/{orgaId}/members` | JWT |
| `poma orga members remove` | `DELETE /orgas/{orgaId}/members/{accountId}` | JWT |
| `poma orga projects` | `GET /orgas/{orgaId}/projects` | JWT |
| `poma orga invitations invite` | `POST /orgas/{orgaId}/invitations` | JWT |
| `poma orga invitations list` | `GET /orgas/{orgaId}/invitations` | JWT |
| `poma orga invitations cancel` | `DELETE /orgas/{orgaId}/invitations/{invitationId}` | JWT |
| `poma orga invitations resend` | `POST /orgas/{orgaId}/invitations/{invitationId}/resend` | JWT |
| `poma orga accept-invitation` | `GET /invitations/accept?token=…` | JWT |

**Flags**

- `list`: `--name` / `-n` (substring filter), `--page`, `--page-size`
- `create`: `--name` / `-n` (required)
- `get`, `update`, `delete`: `--orga-id` / `-o` (required); `update` also `--name` / `-n` (required)
- `members list`, `members add`, `members remove`: `--orga-id` / `-o` (required); `add` also `--email` / `-e` (required), `--role` / `-r` (`admin` or `member`, required); `remove` also `--account-id` / `-a` (required)
- `projects`: `--orga-id` / `-o` (required)
- `invitations invite`: `--orga-id` / `-o` (required), `--email` / `-e` (required)
- `invitations list`: `--orga-id` / `-o` (required), `--status` / `-s` (`pending`/`accepted`/`cancelled`/`expired`/`all`), `--page`, `--page-size`
- `invitations cancel`, `invitations resend`: `--orga-id` / `-o` (required), `--invitation-id` / `-i` (numeric, required)
- `accept-invitation`: `--token` / `-t` (one-time token from invitation email, required)

**Behavior notes**

- `members add` sends an email address (not account ID) and accepts HTTP 201 (added) or 202 (email not found yet, warning returned).
- `invitations cancel` prints `cancelled` on success; `invitations resend` returns the updated invitation JSON.
- `accept-invitation` does not require an orga-id; the token encodes the invitation. Returns JSON with `orga_id` and `message`.

---

### `project` — project management

| Command | API | Auth |
|---------|-----|------|
| `poma project create` | `POST /projects` | JWT |
| `poma project list` | `GET /projects` | JWT |
| `poma project search` | `GET /projects/search` (JSON body) | JWT |
| `poma project get` | `GET /projects/{projectId}` | JWT |
| `poma project delete` | `DELETE /projects/{projectId}` | JWT |

**Flags**

- `create`: `--name` / `-n` (required), `--product` / `-p` (`primecut` or `grill`, required), `--account-id` / `-a`, `--orga-id` / `-o`
- `list`: no subcommand-specific flags
- `search`: `--account-id` / `-a`, `--orga-id` / `-o`, `--project-id` / `-p` (slug), `--name` / `-n`, `--product` / `-P`
- `get`, `delete`: `--project-id` / `-p` (internal UUID — the `id` field on the `Project` object, required)

**Behavior notes**

- `create` returns `{"project":{…}, "api_key":"…"}`. The `api_key` in this response is a **one-time member project API key** shown only at creation time — store it securely.
- `search` sends filters as a JSON request body despite using `GET` (server design); all filter flags are optional.
- `get` and `delete` take the internal project UUID (the `id` field), not the `project_id` slug.

---

### `cheatsheet` — local cheatsheet generation

| Command | API | Auth |
|---------|-----|------|
| `poma cheatsheet create` | *(none — pure local computation)* | No |

**Flags**

- `--input` / `-i` (required): inline JSON string starting with `{`, or a path to a `.json` file **under CWD**
- `--all`: output all cheatsheets as a JSON array `[{"file_id":"…","content":"…"}]`; default prints the first cheatsheet's plain text content

**Input JSON shape**

```json
{
  "relevant_chunksets": [
    { "file_id": "doc_a", "chunks": [0, 2, 5] }
  ],
  "all_chunks": [
    { "chunk_index": 0, "content": "…", "depth": 0, "file_id": "doc_a" }
  ]
}
```

`file_id` defaults to `"single_doc"` when omitted. `chunks` (list of `chunk_index` integers) is required on every chunkset entry. The algorithm expands each hit to include ancestor chunks (parents by depth) and descendant chunks (strictly deeper), then assembles text with `[…]` markers for non-consecutive gaps.

**Behavior notes**

- No HTTP call is made; no token required.
- The `--input` file path is restricted to CWD (`ValidateInputFilePath`).
- Errors on: empty inputs, missing `chunks` key, duplicate `chunk_index` within a `file_id`, `file_id` in chunksets not present in `all_chunks`.
- Unknown chunk IDs referenced by chunksets (not found in `all_chunks`) are silently ignored.

---

### `health` — service check

| Command | API | Auth |
|---------|-----|------|
| `poma health` | `GET /health` | No |

Uses `--base-url` only.

---

## Shared conventions

### Errors

Commands return `error` from Cobra `RunE`; `main` prints it and exits with code 1. Prefer clear, actionable messages; wrap underlying errors where useful.

### Retries

**Not** implemented in the client. Callers (humans or agents) may retry transient network failures themselves. Do not assume automatic backoff.

### Token precedence

Effective token: explicit `--token`, else JSON `token` from `--json` merge, else `POMA_API_KEY` (applied in `PersistentPreRun` so `--help` does not echo the env value). Use the opaque key from `poma account generate-api-key` for stable automation. **Never** print full JWTs or API keys in agent logs or commit them to repos.

### Credit / quota messaging

If the API returns `403` on ingest or elsewhere, use `poma account my-usage`. Exact numeric limits are **not** encoded in this CLI.

---

## Out of scope (this repository)

The CLI does **not**:

- Unpack POMA archives (`chunks.json`, `chunksets.json`, `assets/`) into individual files — consumers do that downstream. (Cheatsheet generation reads pre-unpacked chunk data supplied as JSON input.)
- Store credentials in `~/.config` or similar (operators use env / flags / `--json`).
- Implement vector DB or RAG logic, billing UI.
- Validate file extensions against an allowlist (server enforces supported types).
