package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// fileConfig holds optional JSON keys; only flags that exist on the invoked command are applied.
type fileConfig struct {
	BaseURL       string `json:"base_url,omitempty"`
	StatusBaseURL string `json:"status_base_url,omitempty"`
	Token         string `json:"token,omitempty"`

	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	Company  string `json:"company,omitempty"`
	Code     string `json:"code,omitempty"`

	File   string `json:"file,omitempty"`
	JobID  string `json:"job_id,omitempty"`
	Output string `json:"output,omitempty"`
}

func readConfigBytes(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if strings.HasPrefix(raw, "{") {
		return []byte(raw), nil
	}
	return os.ReadFile(raw)
}

func parseFileConfig(raw string) (*fileConfig, error) {
	b, err := readConfigBytes(raw)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}
	var cfg fileConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse --json: %w", err)
	}
	return &cfg, nil
}

func mergeConfigIntoFlags(cmd *cobra.Command, cfg *fileConfig) error {
	if cfg == nil {
		return nil
	}
	fs := cmd.Flags()
	if err := setIfUnset(fs, "base-url", cfg.BaseURL); err != nil {
		return err
	}
	if err := setIfUnset(fs, "status-base-url", cfg.StatusBaseURL); err != nil {
		return err
	}
	if err := setIfUnset(fs, "token", cfg.Token); err != nil {
		return err
	}
	if err := setIfUnset(fs, "email", cfg.Email); err != nil {
		return err
	}
	if err := setIfUnset(fs, "username", cfg.Username); err != nil {
		return err
	}
	if err := setIfUnset(fs, "company", cfg.Company); err != nil {
		return err
	}
	if err := setIfUnset(fs, "code", cfg.Code); err != nil {
		return err
	}
	if err := setIfUnset(fs, "file", cfg.File); err != nil {
		return err
	}
	if err := setIfUnset(fs, "job-id", cfg.JobID); err != nil {
		return err
	}
	if err := setIfUnset(fs, "output", cfg.Output); err != nil {
		return err
	}
	return nil
}

func setIfUnset(fs *pflag.FlagSet, name, value string) error {
	if value == "" {
		return nil
	}
	flg := fs.Lookup(name)
	if flg == nil || flg.Changed {
		return nil
	}
	return fs.Set(name, value)
}
