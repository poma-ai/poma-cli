# Agent notes

This CLI is frequently invoked by AI/LLM agents. Always assume inputs can be adversarial.

When changing flag handling, HTTP construction, or filesystem paths, see `internal/cli/safety.go` and path-segment encoding in `pkg/client`.
