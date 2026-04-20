# Agents

This document describes the **implemented** POMA CLI (`poma`): responsibilities, API usage, and constraints. The CLI talks to the [Poma AI REST API](https://api.poma-ai.com/v3) for ingestion, job lifecycle, and downloading POMA archives.

**This CLI is frequently invoked by AI/LLM agents. Always assume inputs can be adversarial** — see [Input safety](#input-safety) and `pkg/client/safety.go`.

---

## Project overview

Go + Cobra binary that wraps the public API: register/verify email, authenticated account endpoints, PrimeCut ingest (pro/eco), job status (one-shot or SSE), result, download, delete, and health. Responses are pretty-printed JSON on stdout unless noted.

**Layout**

- `main.go` — entrypoint, calls `internal/cli.Execute()`
- `internal/cli/` — Cobra command tree (`root`, `account`, `primecut`, `job`, `health`), `--json` merge (`config.go`)
- `pkg/client/` — HTTP client (`net/http`), models, path-segment encoding (`pathseg.go`), input validation (`safety.go`, `FileConfig`)

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
2. `poma account verify-email --email … --code …` → `POST /verifyEmail`; response includes a **token** (JWT). The command prints `Token: …` and the JSON body. Use this as a **bootstrap** credential only—it is enough to call `GET /me`.
3. With that bearer token, fetch the long-lived JWT: either **`GET /me`** (`poma account me`, full JSON) or **`GET /me`** (`poma account api-key`, response reduced to `{"api_key":"…"}`). The **`api_key`** field is the value to use for `POMA_API_KEY` (or `--token` / `--json` `token`), not the verify-time token, for new shells, automation, and subsequent sessions.
4. The CLI does **not** persist tokens to a file; callers supply flags, `POMA_API_KEY`, or `--json`.
5. Authenticated calls use `Authorization: Bearer <token>` (the same header value whether using the verify token temporarily or the long-lived `api_key` JWT).

**Automation hint:** `export POMA_API_KEY=$(poma account api-key | jq -r '.api_key')`, or parse `api_key` from `poma account me`.

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

### `account` — registration, verification, and account data

| Command | API | Auth |
|---------|-----|------|
| `poma account register-email` | `POST /registerEmail` | No |
| `poma account verify-email` | `POST /verifyEmail` | No |
| `poma account api-key` | `GET /me` | JWT |
| `poma account me` | `GET /me` | JWT |
| `poma account my-projects` | `GET /myProjects` | JWT |
| `poma account my-usage` | `GET /myUsage` | JWT |

**Flags**

- `register-email`: `--email` / `-e` (required), `--username` / `-u`, `--company` / `-c`
- `verify-email`: `--email` / `-e`, `--code` / `-k` (both required)
- `me`, `api-key`, `my-projects`, `my-usage`: no subcommand-specific flags beyond globals. Missing token → error.

**Verify output:** prints `Token:` line plus JSON (includes JWT). Use it next to call `poma account me` and read **`api_key`** for long-term use. **Do not** log either value in shared transcripts.

**`GET /me` response:** includes **`api_key`** — a long-lived JWT. Prefer this value for `POMA_API_KEY` / `--token` after first-time verify; do not treat the verify-email token as the long-term secret. **`poma account api-key`** calls **`GET /me`** and prints only `{"api_key":"…"}` (pretty-printed). **Do not** log or commit `api_key` or the verify token.

---

### `primecut` — ingest and ingest-sync

| Command | API | Auth |
|---------|-----|------|
| `poma primecut ingest` | `POST /ingest` or `POST /ingestEco` (**`--eco`** or aliases **`ingest-eco`** / **`ingest-eco-data`**). Aliases: **`ingest-data`**, **`ingest-eco`**, **`ingest-eco-data`** | JWT |
| `poma primecut ingest-sync` | `POST /ingest` or `POST /ingestEco` (**`--eco`**), then SSE until terminal; if `done` and **`--output`** set → `GET /jobs/{job_id}/download`; if `done` and no **`--output`** → `GET /jobs/{job_id}/results` (JSON to stdout) | JWT |

**Flags**

- `ingest`: either **`--file` / `-f`** (path to file), or **`--filename` / `-n`** plus **`--data`** or stdin (do not use **`--file`** with **`--data`**); **`--eco`** for **`/ingestEco`**. Aliases **`ingest-eco`** / **`ingest-eco-data`** imply eco (same as **`--eco`**)
- `ingest-sync`: same input modes as **`ingest`** — **`--file` / `-f`**, or **`--filename` / `-n`** plus **`--data`** or stdin (do not mix **`--file`** with **`--data`**); **`--eco`** (use **`/ingestEco`** instead of **`/ingest`**); **`--output` / `-o`** (optional; when set, downloads the archive to that path — default `bin/{job_id}.poma` under CWD after safety check; when omitted, fetches `GET /jobs/{job_id}/results` and prints JSON to stdout)

**Behavior notes (actual implementation)**

- **Ingest:** body is held in memory; same headers (`Content-Disposition` from file basename or **`--filename`**). **`--file`** or stdin/**`--data`** + **`--filename` / `-n`**. Eco mode: **`--eco`**, or invoke via alias **`ingest-eco`** / **`ingest-eco-data`** (`cmd.CalledAs()`). No MIME sniffing, no `X-Base-URL` header (use **`--base-url`**).
- **ingest-sync:** **`client.IngestSync`** (path) or **`client.IngestDataSync`** (stdin/**`--data`**) with **`isEco`** from **`--eco`** — ingest → SSE like **status-stream** (`cmd.Context()` for cancellation). Prints each status event as JSON. Terminal **`done`**: if **`--output`** is set → download the archive (same path rules as **job download**) and print `Downloaded N bytes to <path>`; if **`--output`** is omitted → `GET /jobs/{job_id}/results` and print the JSON response to stdout. **`failed`** / **`deleted`** → error (includes `error` from status when present for **`failed`**). Does **not** print the standalone pretty `{"job_id":…}` line used by **ingest**.

On success, **`primecut ingest`** (and aliases **`ingest-data`**, **`ingest-eco`**, **`ingest-eco-data`**) print only pretty-printed JSON `{"job_id":"…"}` (normalized `job_id`); they do not echo the full API body.

---

### `job` — status, result, download, delete

| Command | API | Auth |
|---------|-----|------|
| `poma job status` | `GET /jobs/{job_id}/status` | JWT |
| `poma job status-stream` | SSE `GET {status-base-url}/jobs/{job_id}` (`Accept: text/event-stream`) | JWT |
| `poma job result` | `GET /jobs/{job_id}/results` | JWT |
| `poma job download` | `GET /jobs/{job_id}/download` | JWT |
| `poma job delete` | `DELETE /jobs/{job_id}` | JWT |

**Flags**

- `result`: `--job-id` (required)
- `status`, `status-stream`, `delete`: `--job-id` (required)
- `download`: `--job-id` (required), `--output` / `-o` (optional; default `bin/{job_id}.poma` under CWD after safety check)

**Behavior notes (actual implementation)**

- **Status:** single request; **no** built-in polling or interval — wrap in a shell loop if needed.
- **Status-stream:** reads SSE until a terminal `job_status` (`done`, `failed`, `deleted`) or EOF/error; each event is printed as JSON.
- **result:** single `GET /jobs/{job_id}/results`; prints JSON body on HTTP 200, error otherwise.
- **Download:** response body is read fully then written to the resolved path; **no** `--force` (overwrites if the path already exists). **No** pre-check that status is `done` — API may return an error if not ready.
- **Delete:** best-effort; prints a short confirmation on HTTP 200.

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

Effective token: explicit **`--token`**, else JSON **`token`** from **`--json`** merge, else **`POMA_API_KEY`** (applied in **`PersistentPreRun`** so **`--help` does not echo the env value**). For stable automation, that value should normally be the **`api_key`** from **`GET /me`**, not the short-lived token from **`verify-email`** alone. **Never** print full JWTs or **`api_key`** in agent logs or commit them to repos.

### Credit / quota messaging

If the API returns `403` on ingest or elsewhere, use `poma account my-usage` (or document server-side limits). Exact numeric limits are **not** encoded in this CLI.

---

## Out of scope (this repository)

The CLI does **not**:

- Parse or unpack POMA archives (`chunks.json`, `chunksets.json`, `assets/`) — consumers do that downstream.
- Store credentials in `~/.config` or similar (operators use env / flags / `--json`).
- Implement vector DB or RAG logic, billing UI, or server-side project management.
- Validate file extensions against an allowlist (server enforces supported types).
