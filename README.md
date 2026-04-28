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

Example: ingest a file, wait until the job finishes, then retrieve the result. Ingest is asynchronous, so you poll or stream status before fetching results.

```bash
# export POMA_API_KEY='<your-jwt>'

# Ingest and receive JSON result on stdout (no --output)
poma primecut ingest-sync --file document.pdf
# poma primecut ingest-sync --filename document.pdf < document.pdf

# Ingest and download the archive instead
poma primecut ingest-sync --file document.pdf --output result.poma

### Alternative: process each step manually

# 1. Submit a file for processing; save the job_id from the output
poma primecut ingest --file document.pdf
# Or pipe bytes: poma primecut ingest --filename document.pdf < document.pdf

# 2. Wait until the job completes (or fails)
poma job status-stream --job-id <job_id>
# Or poll: poma job status --job-id <job_id>

# 3a. When status is "done", fetch the result JSON
poma job result --job-id <job_id>

# 3b. Or download the archive
poma job download --job-id <job_id> --output result.poma

```

**Global flags** (apply to all subcommands):

- `--base-url` — REST API base URL (default: `https://api.poma-ai.com/v3`)
- `--status-base-url` — status / SSE base URL (default: `https://api.poma-ai.com/status/v1`)
- `--token` or env `POMA_API_KEY` — JWT for authenticated requests
- `--json` *(optional)* — merge options from JSON: either an inline object (must start with `{`) or a path to a `.json` file **in the current working directory**. Keys are **snake_case** (e.g. `token`, `job_id`, `file`, `output`, `base_url`). Explicit flags **override** values from `--json`.

### API key

Register for free and try out our ingestion / chunking solution (1000 pages / 100k tokens).

**Via the web app**

1. Sign up at [app.poma-ai.com](https://app.poma-ai.com).
2. Copy your API key from the app and set `POMA_API_KEY`, or pass it with `--token`.

**Via the CLI**

```bash
# 1. Start registration (no token required)
poma account register-email --email you@example.com

# 2. Complete verification with the code from email; the command prints a JWT you can use immediately
poma account verify-email --email you@example.com --code 123456
export POMA_API_KEY='<jwt-from-verify-output>'

# 3. Generate a long-lived opaque API key (recommended for ongoing use)
export POMA_API_KEY=$(poma account generate-api-key | jq -r '.api_key')
```

`poma account generate-api-key` calls `POST /generateApiKey` and prints `{"api_key":"…"}`. Use this key as `POMA_API_KEY` for all subsequent sessions — it is longer-lived and more suitable for automation than the short-lived verify token.

