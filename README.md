# POMA CLI

CLI for the POMA AI public API (v2).

## Install

Requires Go 1.21 or later.

```bash
go install github.com/poma-ai/poma-cli@latest
```

Ensure `$GOPATH/bin` or `$GOBIN` is on your `PATH`. The binary is named `poma`.

To build from source instead:

```bash
git clone https://github.com/poma-ai/poma-cli
cd poma-cli
go mod tidy
go build -o poma .
```

## Project structure

Follows standard Go + Cobra layout:

```
.
├── main.go                 # Entry point (thin)
├── go.mod
├── internal/
│   └── cli/               # Cobra commands
│       ├── root.go        # Root command and global flags
│       ├── user.go        # user subcommands
│       ├── account.go     # account subcommands
│       ├── jobs.go        # jobs subcommands
│       ├── health.go      # health command
│       └── util.go        # shared helpers
└── pkg/
    └── client/            # API client and request/response models
        ├── client.go
        └── models.go
```

## Usage

- `--base-url`: API base URL (default: `https://api.poma-ai.com/v2`)
- `--status-base-url`: Status SSE API base URL (default: `https://api.poma-ai.com/status/v1`)
- `--token` or `POMA_API_TOKEN`: JWT for authenticated endpoints

### Simple flow

1. **Register** – send verification email to your address.
2. **Verify** – use the code from the email to get a JWT.
3. **Ingest** – upload a file; the response gives a `job_id`.
4. **Watch status** – stream job status via SSE until it reaches `done` (or `failed`).
5. **Download** – when status is `done`, download the result.

```bash
# 1. Register (no token)
poma user register-email --email you@example.com

# 2. Verify with code from email; save the token
poma user verify-email --email you@example.com --code 123456
export POMA_API_TOKEN='<token from output>'

# 3. Ingest a file
poma jobs ingest --file document.pdf
# note the job_id from the output

# 4. Stream status until done (or failed)
poma jobs status-stream --job-id <job_id>

# 5. When status is done, download the result
poma jobs download --job-id <job_id> --output result.poma
```

