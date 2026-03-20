package client

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func chdirTemp(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
}

func TestRejectControlChars(t *testing.T) {
	if err := rejectControlChars("ok", "f"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if err := rejectControlChars("a\x00b", "f"); err == nil {
		t.Fatal("expected error for NUL")
	}
	if err := rejectControlChars("a\nb", "f"); err == nil {
		t.Fatal("expected error for newline")
	}
}

func TestRejectJSONInlineC0(t *testing.T) {
	if err := RejectJSONInlineC0("{\n\t\r}"); err != nil {
		t.Fatalf("tab/LF/CR allowed: %v", err)
	}
	if err := RejectJSONInlineC0("a\x00b"); err == nil {
		t.Fatal("expected error for NUL")
	}
	if err := RejectJSONInlineC0("a\x0bb"); err == nil {
		t.Fatal("expected error for vertical tab")
	}
}

func TestValidateResourceName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr string
	}{
		{"empty", "", "required"},
		{"ok", "job-abc-123", ""},
		{"question", "a?b", "?"},
		{"hash", "a#b", "#"},
		{"percent", "a%2f", "%"},
		{"slash", "a/b", "separator"},
		{"backslash", `a\b`, "separator"},
		{"control", "a\x01b", "control"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceName(tt.in, "field")
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err %q should mention %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateHTTPOrigin(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr string
	}{
		{"empty", "", ""},
		{"https_ok", "https://api.example.com/v2", ""},
		{"http_ok", "http://localhost:8080", ""},
		{"no_scheme", "api.example.com", "scheme"},
		{"bad_scheme", "ftp://x.com", "scheme"},
		{"no_host", "https://", "host"},
		{"control", "https://exa\x00mple.com", "control"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHTTPOrigin(tt.raw, "u")
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(strings.ToLower(err.Error()), tt.wantErr) {
				t.Fatalf("err %q should mention %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateJobID(t *testing.T) {
	if err := ValidateJobID("valid-id-1"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateJobID("bad/id"); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateIngestFilePath(t *testing.T) {
	if err := ValidateIngestFilePath(""); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("got %v", err)
	}
	if err := ValidateIngestFilePath("/tmp/ok.pdf"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateIngestFilePath("/tmp/bad\x00.pdf"); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateUserStrings(t *testing.T) {
	if err := ValidateUserStrings("a@b.c", "u", "c", "123"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateUserStrings("ok", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := ValidateUserStrings("x\x00", "", "", ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateFileConfig(t *testing.T) {
	if err := ValidateFileConfig(nil); err != nil {
		t.Fatal(err)
	}
	if err := ValidateFileConfig(&FileConfig{}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateFileConfig(&FileConfig{BaseURL: "ftp://x"}); err == nil {
		t.Fatal("expected bad base_url error")
	}
	if err := ValidateFileConfig(&FileConfig{JobID: "a/b"}); err == nil {
		t.Fatal("expected bad job_id")
	}
}

func TestValidateSafeOutputDir_underCWD(t *testing.T) {
	chdirTemp(t)

	got, err := ValidateSafeOutputDir(filepath.Join("bin", "out.poma"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(got, filepath.Join("bin", "out.poma")) {
		t.Fatalf("got %q", got)
	}
}

func TestValidateSafeOutputDir_rejectsEscape(t *testing.T) {
	chdirTemp(t)

	if _, err := ValidateSafeOutputDir(".."); err == nil {
		t.Fatal("expected escape error")
	}
}

func TestValidateJSONFilePathUnderCwd_rejectsEscape(t *testing.T) {
	chdirTemp(t)

	if _, err := ValidateJSONFilePathUnderCwd(".."); err == nil {
		t.Fatal("expected escape error")
	}
}

func TestValidatePersistentFlags(t *testing.T) {
	if err := ValidatePersistentFlags("https://example.com", "https://status.example.com", "", ""); err != nil {
		t.Fatal(err)
	}

	if err := ValidatePersistentFlags("https://example.com", "https://status.example.com", "tok\x00en", ""); err == nil {
		t.Fatal("expected token control char error")
	}

	if err := ValidatePersistentFlags("https://example.com", "https://status.example.com", "", "{bad\x00json"); err == nil {
		t.Fatal("expected inline json error")
	}
}
