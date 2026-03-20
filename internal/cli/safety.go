package cli

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// reject_control_chars rejects ASCII C0 controls (U+0000–U+001F).
func reject_control_chars(s string, field string) error {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 {
			return fmt.Errorf("%s: contains control character (0x%02x at byte %d)", field, s[i], i)
		}
	}
	return nil
}

// reject_json_inline_c0 rejects C0 controls except tab, LF, CR so inline --json can be pretty-printed.
func reject_json_inline_c0(s string) error {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 0x20 && b != '\t' && b != '\n' && b != '\r' {
			return fmt.Errorf("--json: contains disallowed control character (0x%02x at byte %d)", b, i)
		}
	}
	return nil
}

// validate_resource_name rejects characters that break URL paths or signal query/fragments or pre-encoding.
func validate_resource_name(s, field string) error {
	if s == "" {
		return fmt.Errorf("%s: required", field)
	}
	if strings.ContainsAny(s, "?#%") {
		return fmt.Errorf("%s: must not contain '?', '#', or '%%'", field)
	}
	if strings.ContainsAny(s, `/\`) {
		return fmt.Errorf("%s: must not contain path separators", field)
	}
	return reject_control_chars(s, field)
}

// validate_safe_output_dir returns a canonical absolute path for writes, confined under the process CWD.
func validate_safe_output_dir(outPath string) (string, error) {
	if err := reject_control_chars(outPath, "output path"); err != nil {
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

// validate_json_file_path_under_cwd resolves a filesystem path for --json and ensures it stays under CWD.
func validate_json_file_path_under_cwd(path string) (string, error) {
	if err := reject_control_chars(path, "--json file path"); err != nil {
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

func validate_http_origin(raw, field string) error {
	if raw == "" {
		return nil
	}
	if err := reject_control_chars(raw, field); err != nil {
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

func validate_persistent_flags() error {
	if err := validate_http_origin(baseURL, "--base-url"); err != nil {
		return err
	}
	if err := validate_http_origin(statusBaseURL, "--status-base-url"); err != nil {
		return err
	}
	if token != "" {
		if err := reject_control_chars(token, "--token"); err != nil {
			return err
		}
	}
	if jsonArg != "" {
		t := strings.TrimSpace(jsonArg)
		if strings.HasPrefix(t, "{") {
			if err := reject_json_inline_c0(jsonArg); err != nil {
				return err
			}
		} else {
			if _, err := validate_json_file_path_under_cwd(t); err != nil {
				return err
			}
		}
	}
	return nil
}

func validate_job_id(jobID string) error {
	return validate_resource_name(jobID, "--job-id")
}

func validate_ingest_file_path(path string) error {
	if path == "" {
		return fmt.Errorf("--file is required")
	}
	return reject_control_chars(path, "--file")
}

func validate_user_strings(email, username, company, code string) error {
	if email != "" {
		if err := reject_control_chars(email, "--email"); err != nil {
			return err
		}
	}
	if username != "" {
		if err := reject_control_chars(username, "--username"); err != nil {
			return err
		}
	}
	if company != "" {
		if err := reject_control_chars(company, "--company"); err != nil {
			return err
		}
	}
	if code != "" {
		if err := reject_control_chars(code, "--code"); err != nil {
			return err
		}
	}
	return nil
}

func validate_file_config(cfg *fileConfig) error {
	if cfg == nil {
		return nil
	}
	if err := validate_http_origin(cfg.BaseURL, "json base_url"); err != nil {
		return err
	}
	if err := validate_http_origin(cfg.StatusBaseURL, "json status_base_url"); err != nil {
		return err
	}
	if err := reject_control_chars(cfg.Token, "json token"); err != nil {
		return err
	}
	if err := reject_control_chars(cfg.Email, "json email"); err != nil {
		return err
	}
	if err := reject_control_chars(cfg.Username, "json username"); err != nil {
		return err
	}
	if err := reject_control_chars(cfg.Company, "json company"); err != nil {
		return err
	}
	if err := reject_control_chars(cfg.Code, "json code"); err != nil {
		return err
	}
	if err := reject_control_chars(cfg.File, "json file"); err != nil {
		return err
	}
	if cfg.JobID != "" {
		if err := validate_resource_name(cfg.JobID, "json job_id"); err != nil {
			return err
		}
	}
	if cfg.Output != "" {
		if _, err := validate_safe_output_dir(cfg.Output); err != nil {
			return err
		}
	}
	return nil
}
