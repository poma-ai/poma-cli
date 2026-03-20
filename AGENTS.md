# Agents

This document describes the **implemented** POMA CLI (`poma`): responsibilities, API usage, and constraints. The CLI talks to the [Poma AI REST API](https://api.poma-ai.com/v2) for ingestion, job lifecycle, and downloading POMA archives.

**This CLI is frequently invoked by AI/LLM agents. Always assume inputs can be adversarial** — see [Input safety](#input-safety) and `pkg/client/safety.go`.

---

## Project overview

Go + Cobra binary that wraps the public API: register/verify email, authenticated account endpoints, job ingest (pro/eco), job status (one-shot or SSE), download, delete, and health. Responses are pretty-printed JSON on stdout unless noted.

**Layout**

- `main.go` — entrypoint, calls `internal/cli.Execute()`
- `internal/cli/` — Cobra commands, `--json` merge (`config.go`)
- `pkg/client/` — HTTP client (`net/http`), models, path-segment encoding (`pathseg.go`), input validation (`safety.go`, `FileConfig`)

---

## Stack & constraints

| Area | Implementation |
|------|----------------|
| **Language** | Go 1.21+ (`go.mod`) |
| **CLI** | [Cobra](https://github.com/spf13/cobra) |
| **HTTP** | `net/http`; `Authorization: Bearer <token>` when `Client.Token` is set |
| **Config / secrets** | Flags, `POMA_API_TOKEN`, and optional `--json` (see below). **No** on-disk config file in this repo. |
| **Output** | `fmt` + `PrintJSON` for API bodies; errors to stderr via caller (`main` prints error and exits 1) |

---

## Global flags (all subcommands)

| Flag | Default | Notes |
|------|---------|--------|
| `--base-url` | `https://api.poma-ai.com/v2` | Must be `http` or `https` with a host. |
| `--status-base-url` | `https://api.poma-ai.com/status/v1` | Used by `jobs status-stream` only. |
| `--token` | `$POMA_API_TOKEN` | Required for authenticated routes. |
| `--json` | (empty) | Inline JSON object (`{...}`) **or** path to a `.json` file under the process **CWD**; keys map to flags; explicit flags **override** JSON. |

### `--json` keys (snake_case)

`base_url`, `status_base_url`, `token`, `email`, `username`, `company`, `code`, `file`, `job_id`, `output` — only flags that exist on the invoked command are applied.

---

## Authentication flow

1. `poma user register-email --email …` → `POST /registerEmail` (optional `--username`, `--company`). No JWT.
2. `poma user verify-email --email … --code …` → `POST /verifyEmail`; response includes a **token** (JWT). The command prints `Token: …` and the JSON body. Use this as a **bootstrap** credential only—it is enough to call `GET /me`.
3. With that bearer token, fetch the long-lived JWT: either **`GET /me`** (`poma account me`, full JSON) or **`GET /me`** (`poma account api-key`, response reduced to `{"api_key":"…"}`). The **`api_key`** field is the value to use for `POMA_API_TOKEN` (or `--token` / `--json` `token`), not the verify-time token, for new shells, automation, and subsequent sessions.
4. The CLI does **not** persist tokens to a file; callers supply flags, `POMA_API_TOKEN`, or `--json`.
5. Authenticated calls use `Authorization: Bearer <token>` (the same header value whether using the verify token temporarily or the long-lived `api_key` JWT).

**Automation hint:** `export POMA_API_TOKEN=$(poma account api-key | jq -r '.api_key')`, or parse `api_key` from `poma account me`.

---

## Input safety

Implemented in `pkg/client/safety.go` and `pkg/client` (path segments, `Content-Disposition` filename):

- **C0 controls** (ASCII `0x00`–`0x1F`): rejected on most flag/JSON strings; inline `--json` allows tab/LF/CR only.
- **Job IDs** (`--job-id`, JSON `job_id`): no `?` `#` `%`, no path separators; must be a single path segment for URLs.
- **Download output** (`--output` / JSON `output`): resolved and **must stay under the process current working directory** (same rule for `--json` when it is a **file path**).
- **HTTP paths**: job IDs are passed through `url.PathEscape` in `pkg/client` when building `/jobs/...` URLs.

Ingest `--file` rejects control characters but **may** be any readable path (not restricted to CWD).

---

## Command reference (by area)

Below, “agent” names are **logical groupings** for automation docs; each maps to real `poma …` subcommands.

### `user` — registration & verification

| Command | API | Auth |
|---------|-----|------|
| `poma user register-email` | `POST /registerEmail` | No |
| `poma user verify-email` | `POST /verifyEmail` | No |

**Flags**

- `register-email`: `--email` / `-e` (required), `--username` / `-u`, `--company` / `-c`
- `verify-email`: `--email` / `-e`, `--code` / `-k` (both required)

**Verify output:** prints `Token:` line plus JSON (includes JWT). Use it next to call `poma account me` and read **`api_key`** for long-term use. **Do not** log either value in shared transcripts.

---

### `account` — authenticated account data

| Command | API | Auth |
|---------|-----|------|
| `poma account me` | `GET /me` | JWT |
| `poma account api-key` | `GET /me` | JWT |
| `poma account my-projects` | `GET /myProjects` | JWT |
| `poma account my-usage` | `GET /myUsage` | JWT |

No subcommand-specific flags beyond globals. Missing token → error.

**`GET /me` response:** includes **`api_key`** — a long-lived JWT. Prefer this value for `POMA_API_TOKEN` / `--token` after first-time verify; do not treat the verify-email token as the long-term secret. **`poma account api-key`** calls **`GET /me`** and prints only `{"api_key":"…"}` (pretty-printed). **Do not** log or commit `api_key` or the verify token.

---

### `jobs` — ingest, status, download, delete

| Command | API | Auth |
|---------|-----|------|
| `poma jobs ingest` | `POST /ingest` (raw body, `application/octet-stream`) | JWT |
| `poma jobs ingest-eco` | `POST /ingestEco` (same shape) | JWT |
| `poma jobs status` | `GET /jobs/{job_id}/status` | JWT |
| `poma jobs status-stream` | SSE `GET {status-base-url}/jobs/{job_id}` (`Accept: text/event-stream`) | JWT |
| `poma jobs download` | `GET /jobs/{job_id}/download` | JWT |
| `poma jobs delete` | `DELETE /jobs/{job_id}` | JWT |

**Flags**

- `ingest`, `ingest-eco`: `--file` / `-f` (required)
- `status`, `status-stream`, `delete`: `--job-id` (required)
- `download`: `--job-id` (required), `--output` / `-o` (optional; default `bin/{job_id}.poma` under CWD after safety check)

**Behavior notes (actual implementation)**

- **Ingest:** reads the **entire** file into memory, sets `Content-Disposition` (sanitized basename), `Content-Type: application/octet-stream`, `Content-Length`. No MIME sniffing, no `X-Base-URL` header (use `--base-url`).
- **Status:** single request; **no** built-in polling or interval — wrap in a shell loop if needed.
- **Status-stream:** reads SSE until a terminal `job_status` (`done`, `failed`, `deleted`) or EOF/error; each event is printed as JSON.
- **Download:** response body is read fully then written to the resolved path; **no** `--force` (overwrites if the path already exists). **No** pre-check that status is `done` — API may return an error if not ready.
- **Delete:** best-effort; prints a short confirmation on HTTP 200.

On success, **`jobs ingest`** and **`jobs ingest-eco`** print only pretty-printed JSON `{"job_id":"…"}` (normalized `job_id`); they do not echo the full API body.

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

Effective token: `--token` / JSON `token` after merge, else default from `POMA_API_TOKEN` at flag definition time. For stable automation, that value should normally be the **`api_key`** from `GET /me`, not the short-lived token from `verify-email` alone. **Never** print full JWTs or `api_key` in agent logs or commit them to repos.

### Credit / quota messaging

If the API returns `403` on ingest or elsewhere, use `poma account my-usage` (or document server-side limits). Exact numeric limits are **not** encoded in this CLI.

---

## Out of scope (this repository)

The CLI does **not**:

- Parse or unpack POMA archives (`chunks.json`, `chunksets.json`, `assets/`) — consumers do that downstream.
- Store credentials in `~/.config` or similar (operators use env / flags / `--json`).
- Implement vector DB or RAG logic, billing UI, or server-side project management.
- Validate file extensions against an allowlist (server enforces supported types).
