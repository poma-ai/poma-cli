package client

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// FileConfig holds optional --json keys (snake_case in JSON). Only flags that exist on the
// invoked command are applied by the CLI.
type FileConfig struct {
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

func rejectControlChars(s string, field string) error {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 {
			return fmt.Errorf("%s: contains control character (0x%02x at byte %d)", field, s[i], i)
		}
	}
	return nil
}

// RejectJSONInlineC0 rejects C0 controls except tab, LF, CR so inline --json can be pretty-printed.
func RejectJSONInlineC0(s string) error {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 0x20 && b != '\t' && b != '\n' && b != '\r' {
			return fmt.Errorf("--json: contains disallowed control character (0x%02x at byte %d)", b, i)
		}
	}
	return nil
}

// ValidateResourceName rejects characters that break URL paths or signal query/fragments or pre-encoding.
func ValidateResourceName(s, field string) error {
	if s == "" {
		return fmt.Errorf("%s: required", field)
	}
	if strings.ContainsAny(s, "?#%") {
		return fmt.Errorf("%s: must not contain '?', '#', or '%%'", field)
	}
	if strings.ContainsAny(s, `/\`) {
		return fmt.Errorf("%s: must not contain path separators", field)
	}
	return rejectControlChars(s, field)
}

// ValidateSafeOutputDir returns a canonical absolute path for writes, confined under the process CWD.
func ValidateSafeOutputDir(outPath string) (string, error) {
	if err := rejectControlChars(outPath, "output path"); err != nil {
		return "", err
	}
	wdAbs, err := absCanonicalWD()
	if err != nil {
		return "", err
	}
	var target string
	if filepath.IsAbs(outPath) {
		target = filepath.Clean(outPath)
	} else {
		target = filepath.Clean(filepath.Join(wdAbs, outPath))
	}
	rel, err := filepath.Rel(wdAbs, target)
	if err != nil {
		return "", fmt.Errorf("output path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("output path escapes working directory")
	}
	return target, nil
}

// ValidateJSONFilePathUnderCwd resolves a filesystem path for --json and ensures it stays under CWD.
func ValidateJSONFilePathUnderCwd(path string) (string, error) {
	if err := rejectControlChars(path, "--json file path"); err != nil {
		return "", err
	}
	wdAbs, err := absCanonicalWD()
	if err != nil {
		return "", err
	}
	var target string
	if filepath.IsAbs(path) {
		target = filepath.Clean(path)
	} else {
		target = filepath.Clean(filepath.Join(wdAbs, path))
	}
	rel, err := filepath.Rel(wdAbs, target)
	if err != nil {
		return "", fmt.Errorf("--json file path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("--json file path escapes working directory")
	}
	return target, nil
}

func absCanonicalWD() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	a, err := filepath.Abs(wd)
	if err != nil {
		return "", err
	}
	a, err = filepath.EvalSymlinks(a)
	if err != nil {
		return "", err
	}
	return filepath.Clean(a), nil
}

// ValidateHTTPOrigin checks non-empty strings are http(s) URLs with a host; empty is allowed.
func ValidateHTTPOrigin(raw, field string) error {
	if raw == "" {
		return nil
	}
	if err := rejectControlChars(raw, field); err != nil {
		return err
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%s: invalid URL: %w", field, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%s: URL must use http or https scheme", field)
	}
	if u.Host == "" {
		return fmt.Errorf("%s: URL must include a host", field)
	}
	return nil
}

// ValidatePersistentFlags validates global CLI strings used before each command runs.
func ValidatePersistentFlags(baseURL, statusBaseURL, token, jsonArg string) error {
	if err := ValidateHTTPOrigin(baseURL, "--base-url"); err != nil {
		return err
	}
	if err := ValidateHTTPOrigin(statusBaseURL, "--status-base-url"); err != nil {
		return err
	}
	if token != "" {
		if err := rejectControlChars(token, "--token"); err != nil {
			return err
		}
	}
	if jsonArg != "" {
		t := strings.TrimSpace(jsonArg)
		if strings.HasPrefix(t, "{") {
			if err := RejectJSONInlineC0(jsonArg); err != nil {
				return err
			}
		} else {
			if _, err := ValidateJSONFilePathUnderCwd(t); err != nil {
				return err
			}
		}
	}
	return nil
}

// ValidateJobID checks --job-id values for safe path segments.
func ValidateJobID(jobID string) error {
	return ValidateResourceName(jobID, "--job-id")
}

// ValidateIngestFilePath rejects control characters in --file paths and checks the file is readable.
func ValidateIngestFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("--file is required")
	}
	if err := rejectControlChars(path, "--file"); err != nil {
		return err
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("--file: %w", err)
	}
	return nil
}

// ValidateUserStrings rejects C0 controls in user registration / verification fields.
func ValidateUserStrings(email, username, company, code string) error {
	if email != "" {
		if err := rejectControlChars(email, "--email"); err != nil {
			return err
		}
	}
	if username != "" {
		if err := rejectControlChars(username, "--username"); err != nil {
			return err
		}
	}
	if company != "" {
		if err := rejectControlChars(company, "--company"); err != nil {
			return err
		}
	}
	if code != "" {
		if err := rejectControlChars(code, "--code"); err != nil {
			return err
		}
	}
	return nil
}

// ValidateFileConfig validates JSON flag overlay values before merge.
func ValidateFileConfig(cfg *FileConfig) error {
	if cfg == nil {
		return nil
	}
	if err := ValidateHTTPOrigin(cfg.BaseURL, "json base_url"); err != nil {
		return err
	}
	if err := ValidateHTTPOrigin(cfg.StatusBaseURL, "json status_base_url"); err != nil {
		return err
	}
	if err := rejectControlChars(cfg.Token, "json token"); err != nil {
		return err
	}
	if err := rejectControlChars(cfg.Email, "json email"); err != nil {
		return err
	}
	if err := rejectControlChars(cfg.Username, "json username"); err != nil {
		return err
	}
	if err := rejectControlChars(cfg.Company, "json company"); err != nil {
		return err
	}
	if err := rejectControlChars(cfg.Code, "json code"); err != nil {
		return err
	}
	if err := rejectControlChars(cfg.File, "json file"); err != nil {
		return err
	}
	if cfg.JobID != "" {
		if err := ValidateResourceName(cfg.JobID, "json job_id"); err != nil {
			return err
		}
	}
	if cfg.Output != "" {
		if _, err := ValidateSafeOutputDir(cfg.Output); err != nil {
			return err
		}
	}
	return nil
}
