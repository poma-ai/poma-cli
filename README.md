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

You can install the `poma` binary in any of these ways:

**Homebrew** — [Homebrew](https://brew.sh) serves prebuilt releases from the [`poma-ai/poma`](https://github.com/poma-ai/homebrew-poma) tap:

```bash
brew tap poma-ai/poma
brew install poma
```

**Go toolchain** — requires Go 1.21 or later:

```bash
go install github.com/poma-ai/poma-cli@latest
```

Put `$GOPATH/bin` or `$GOBIN` on your `PATH` if it is not already.

**From source** — clone this repository and build (Go 1.21+):

```bash
git clone https://github.com/poma-ai/poma-cli
cd poma-cli
go mod tidy
go build -o poma .
```

## Usage

Most API calls need a JWT (see [API key](#api-key) below). Export it or pass `--token` on each command.

Example: ingest a file, wait until the job finishes, then download the result. Ingest is asynchronous, so you poll or stream status before downloading.

```bash
# export POMA_API_TOKEN='<your-jwt>'

# 1. Submit a file for processing; save the job_id from the output
poma jobs ingest --file document.pdf

# 2. Wait until the job completes (or fails)
poma jobs status-stream --job-id <job_id>
# Or poll: poma jobs status --job-id <job_id>

# 3. When status is "done", fetch the artifact
poma jobs download --job-id <job_id> --output result.poma
```

**Global flags** (apply to all subcommands):

- `--base-url` — REST API base URL (default: `https://api.poma-ai.com/v2`)
- `--status-base-url` — status / SSE base URL (default: `https://api.poma-ai.com/status/v1`)
- `--token` or env `POMA_API_TOKEN` — JWT for authenticated requests
- `--json` *(optional)* — merge options from JSON: either an inline object (must start with `{`) or a path to a `.json` file **in the current working directory**. Keys are **snake_case** (e.g. `token`, `job_id`, `file`, `output`, `base_url`). Explicit flags **override** values from `--json`.

### API key

Register for free and try out our ingestion / chunking solution (1000 pages / 100k tokens).

**Via the web app**

1. Sign up at [app.poma-ai.com](https://app.poma-ai.com).
2. Copy your API key from the app and set `POMA_API_TOKEN`, or pass it with `--token`.

**Via the CLI**

```bash
# 1. Start registration (no token required)
poma user register-email --email you@example.com

# 2. Complete verification with the code from email; the command prints a JWT you can use immediately
poma user verify-email --email you@example.com --code 123456
export POMA_API_TOKEN='<jwt-from-verify-output>'

# 3. Optional: replace with the long-lived JWT from your account (requires jq)
export POMA_API_TOKEN=$(poma account api-key | jq -r '.api_key')
```

`poma account api-key` prints JSON with an `api_key` field (a JWT suitable for ongoing CLI use).

## Project structure

Standard Go layout with Cobra under `internal/cli` and a small HTTP client in `pkg/client`:

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

