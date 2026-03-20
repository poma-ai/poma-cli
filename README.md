![POMA AI Logo](https://raw.githubusercontent.com/poma-ai/.github/main/assets/POMA_AI_Logo_Pink.svg)

[![](https://img.shields.io/badge/patented%20at%20USPTO-8A2BE2)]()
[![](https://img.shields.io/badge/patented%20at%20DPA-8A2BE2)]()
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)](LICENSE)

# POMA CLI

**Problem:**

"Current RAG systems chunk documents the same way a paper shredder does — fast, uniform, and completely unaware of meaning.

Tables get split in half. Headings lose their content. A sentence that only makes sense in context gets retrieved alone. The AI does its best with the fragments it receives — but when the fragments are broken, the answers are too.

Hallucinations, missed facts, and lost context aren't model problems. They're chunking problems."

**Solution:**

POMA AI PrimeCut.

PrimeCut understands your document's content hierarchy before chunking — preserving structural relationships, eliminating context poisoning, and producing semantically coherent chunks that make every downstream RAG component more accurate by default.

## Install

**Homebrew**

Install with [Homebrew](https://brew.sh):

```bash
brew tap poma-ai/poma
brew install poma
```

**Go**

Requires Go 1.21 or later:

```bash
go install github.com/poma-ai/poma-cli@latest
```

Ensure `$GOPATH/bin` or `$GOBIN` is on your `PATH`. The binary is named `poma`.

**Source**

To build from source instead (Go 1.21+):

```bash
git clone https://github.com/poma-ai/poma-cli
cd poma-cli
go mod tidy
go build -o poma .
```

## Usage

- `--base-url`: API base URL (default: `https://api.poma-ai.com/v2`)
- `--status-base-url`: Status SSE API base URL (default: `https://api.poma-ai.com/status/v1`)
- `--token` or `POMA_API_TOKEN`: JWT for authenticated endpoints
- `--json` *(optional)*: pass options as JSON instead of repeating flags. Value is either an inline object (must start with `{`) or a path to a `.json` file **under your current working directory**. Keys use **snake_case** (e.g. `token`, `job_id`, `file`, `output`, `base_url`). Any flag you set on the command line **overrides** the same field from `--json`.

### Simple flow (api key)

Prerequisite: set **`POMA_API_TOKEN`** to your long-lived **`api_key`** (JWT). Copy it from the [POMA web app](https://app.poma-ai.com) after you sign in, or run the **Full flow (incl. registration/login)** section below (`verify-email`, then `account api-key`).

```bash
# export POMA_API_TOKEN='<paste api_key here>'

# 4. Ingest a file
poma jobs ingest --file document.pdf
# note the job_id from the output

# 5. Stream status until done (or failed)
poma jobs status-stream --job-id <job_id>

# 6. When status is done, download the result
poma jobs download --job-id <job_id> --output result.poma
```

### Full flow (incl. registration/login)

1. **Register** – send verification email to your address.
2. **Verify** – use the code from the email; the CLI prints a JWT you can use for the next step.
3. **Long-lived JWT** – with that JWT, use **`poma account api-key`** (`GET /me`; prints only `{"api_key":"…"}`) or **`poma account me`** (`GET /me`, full account JSON). The **`api_key`** value is the long-lived JWT—set `POMA_API_TOKEN` for day-to-day use (new shells, automation, etc.).
4. **Ingest** – upload a file; the response gives a `job_id`.
5. **Watch status** – stream job status via SSE until it reaches `done` (or `failed`).
6. **Download** – when status is `done`, download the result.

```bash
# 1. Register (no token)
poma user register-email --email you@example.com

# 2. Verify with code from email (JWT from verify is enough for account me / api-key)
poma user verify-email --email you@example.com --code 123456
export POMA_API_TOKEN='<JWT from verify output>'

# 3. Long-lived JWT — GET /me, field "api_key"
export POMA_API_TOKEN=$(poma account api-key | jq -r '.api_key')

# 4. Ingest a file
poma jobs ingest --file document.pdf
# note the job_id from the output

# 5. Stream status until done (or failed)
poma jobs status-stream --job-id <job_id>

# 6. When status is done, download the result
poma jobs download --job-id <job_id> --output result.poma
```

## Project structure

Follows standard Go + Cobra layout:

```
.
├── main.go                 # Entry point (thin)
├── go.mod
├── go.sum
├── README.md
├── AGENTS.md               # Full CLI/API reference for agents
├── SKILL.md                # Short agent/dev checklist
├── internal/
│   └── cli/                # Cobra commands
│       ├── root.go         # Root command, global flags, --json hook
│       ├── config.go       # JSON config shape and flag merge
│       ├── user.go         # user subcommands
│       ├── account.go      # account subcommands
│       ├── jobs.go         # jobs subcommands
│       ├── health.go       # health command
│       └── util.go         # shared helpers (e.g. PrintJSON)
└── pkg/
    └── client/             # HTTP API client
        ├── client.go
        ├── models.go       # Request/response types
        ├── pathseg.go      # URL path-segment encoding
        └── safety.go       # Input validation, FileConfig (--json shape)
```

