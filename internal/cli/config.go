package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/poma-ai/poma-cli/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func readConfigBytes(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if strings.HasPrefix(raw, "{") {
		if err := client.RejectJSONInlineC0(raw); err != nil {
			return nil, err
		}
		return []byte(raw), nil
	}
	path, err := client.ValidateJSONFilePathUnderCwd(raw)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

func parseFileConfig(raw string) (*client.FileConfig, error) {
	b, err := readConfigBytes(raw)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}
	var cfg client.FileConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse --json: %w", err)
	}
	return &cfg, nil
}

func mergeConfigIntoFlags(cmd *cobra.Command, cfg *client.FileConfig) error {
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
