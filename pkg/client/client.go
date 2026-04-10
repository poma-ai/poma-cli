package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// Client calls the POMA API v2 (public OpenAPI).
type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

// New returns a client.
func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTP:    &http.Client{},
	}
}

// Do sends an HTTP request and returns the response body and error.
// Caller must not expect body to be valid for non-2xx if err is nil; we still return body on 4xx/5xx.
func (c *Client) Do(method, path string, body io.Reader, headers map[string]string) ([]byte, int, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if c.Token != "" && req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}

// DoJSON sends a JSON request and returns body and status. If body is nil, Content-Type is not set for GET.
func (c *Client) DoJSON(method, path string, reqBody any) ([]byte, int, error) {
	var body io.Reader
	headers := map[string]string{}
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return nil, 0, err
		}
		body = bytes.NewReader(b)
		headers["Content-Type"] = "application/json"
	}
	return c.Do(method, path, body, headers)
}

// RegisterEmail calls POST /registerEmail (no auth).
func (c *Client) RegisterEmail(req *AccountRegisterEmailRequest) ([]byte, int, error) {
	return c.DoJSON(http.MethodPost, "/registerEmail", req)
}

// VerifyEmail calls POST /verifyEmail (no auth). Returns token in response.
func (c *Client) VerifyEmail(req *AccountVerifyEmailRequest) ([]byte, int, error) {
	return c.DoJSON(http.MethodPost, "/verifyEmail", req)
}

// Ingest sends POST /ingest with raw file body (pro).
func (c *Client) Ingest(filePath string) ([]byte, int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, 0, err
	}
	return c.IngestData(data, filepath.Base(filePath))
}

// IngestData sends POST /ingest with raw bytes (pro). filename is the basename used in Content-Disposition only.
func (c *Client) IngestData(data []byte, filename string) ([]byte, int, error) {
	if len(data) == 0 {
		return nil, 0, fmt.Errorf("ingest body is empty")
	}
	name := sanitizeContentDispositionFilename(filepath.Base(filename))
	headers := map[string]string{
		"Content-Disposition": `attachment; filename="` + name + `"`,
		"Content-Type":        "application/octet-stream",
		"Content-Length":      strconv.Itoa(len(data)),
	}
	return c.Do(http.MethodPost, "/ingest", bytes.NewReader(data), headers)
}

// IngestEco sends POST /ingestEco with raw file body (eco).
func (c *Client) IngestEco(filePath string) ([]byte, int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, 0, err
	}
	return c.IngestEcoData(data, filepath.Base(filePath))
}

// IngestEcoData sends POST /ingestEco with raw bytes (eco). filename is the basename used in Content-Disposition only.
func (c *Client) IngestEcoData(data []byte, filename string) ([]byte, int, error) {
	if len(data) == 0 {
		return nil, 0, fmt.Errorf("ingest body is empty")
	}
	name := sanitizeContentDispositionFilename(filepath.Base(filename))
	headers := map[string]string{
		"Content-Disposition": `attachment; filename="` + name + `"`,
		"Content-Type":        "application/octet-stream",
		"Content-Length":      strconv.Itoa(len(data)),
	}
	return c.Do(http.MethodPost, "/ingestEco", bytes.NewReader(data), headers)
}

// IngestSync reads the file into memory, then calls IngestDataSync.
func (c *Client) IngestSync(ctx context.Context, filePath string, isEco bool, statusBaseURL string, resolveOut func(jobID string) (outPath string, err error), onStatus func(*JobStatus)) (written int64, outPath string, retErr error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, "", err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return 0, "", err
	}
	return c.IngestDataSync(ctx, data, filepath.Base(filePath), isEco, statusBaseURL, resolveOut, onStatus)
}

// IngestDataSync POSTs /ingest or /ingestEco with raw bytes, follows the status SSE stream until a
// terminal job_status (done, failed, deleted), then downloads when status is done.
// resolveOut returns a safe local path for DownloadJob; it is only called when the terminal status is done.
// If resolveOut is nil, the download is skipped and (0, jobID, nil) is returned on done.
// onStatus is optional; if set, it is invoked for each parsed job_status event.
// On success with done and resolveOut set, returns bytes written and the path passed to DownloadJob.
func (c *Client) IngestDataSync(ctx context.Context, data []byte, filename string, isEco bool, statusBaseURL string, resolveOut func(jobID string) (outPath string, err error), onStatus func(*JobStatus)) (written int64, outPath string, retErr error) {
	if c.Token == "" {
		return 0, "", fmt.Errorf("token is required")
	}
	var body []byte
	var st int
	var err error
	if isEco {
		body, st, err = c.IngestEcoData(data, filename)
	} else {
		body, st, err = c.IngestData(data, filename)
	}
	PrintIngestJobIDOnly(body)

	if err != nil {
		return 0, "", err
	}
	if st != http.StatusCreated {
		return 0, "", fmt.Errorf("ingest: HTTP %d: %s", st, string(body))
	}
	j, err := ParseJob(body)
	if err != nil {
		return 0, "", fmt.Errorf("parse ingest response: %w", err)
	}
	if j.JobID == "" {
		return 0, "", fmt.Errorf("ingest response has no job_id")
	}
	jobID := j.JobID
	if err = ValidateJobID(jobID); err != nil {
		return 0, "", err
	}

	var last *JobStatus
	streamErr := c.StatusStream(ctx, jobID, statusBaseURL, func(s *JobStatus) bool {
		last = s
		if onStatus != nil {
			onStatus(s)
		}
		return true
	})
	if streamErr != nil {
		return 0, "", streamErr
	}
	if last == nil {
		return 0, "", fmt.Errorf("no job status events received")
	}

	switch last.Status {
	case "done":
		if resolveOut == nil {
			return 0, jobID, nil
		}
		safeOut, err := resolveOut(jobID)
		if err != nil {
			return 0, "", err
		}
		n, dlStatus, err := c.DownloadJob(jobID, safeOut)
		if err != nil {
			return 0, "", err
		}
		if dlStatus != http.StatusOK {
			return 0, "", fmt.Errorf("download: HTTP %d", dlStatus)
		}
		return n, safeOut, nil
	case "failed":
		if last.Error != "" {
			return 0, "", fmt.Errorf("job failed: %s", last.Error)
		}
		return 0, "", fmt.Errorf("job failed")
	case "deleted":
		return 0, "", fmt.Errorf("job deleted")
	default:
		return 0, "", fmt.Errorf("job ended in non-terminal state: %q", last.Status)
	}
}

// GetJobStatus returns GET /jobs/{job_id}/status.
func (c *Client) GetJobStatus(jobID string) ([]byte, int, error) {
	seg := JobPathSegment(jobID)
	return c.Do(http.MethodGet, "/jobs/"+seg+"/status", nil, nil)
}

// StatusStream opens the Status API SSE stream for a job (GET /jobs/{job_id} on status API).
// statusBaseURL is the Status API base, e.g. "https://api.poma-ai.com/status/v1".
// For each job_status event, onEvent is called with the parsed JobStatus; if onEvent returns false, streaming stops.
// The stream ends when the job reaches a terminal state (done, failed, deleted) or the context is cancelled.
func (c *Client) StatusStream(ctx context.Context, jobID, statusBaseURL string, onEvent func(*JobStatus) bool) error {
	reqURL := strings.TrimSuffix(statusBaseURL, "/") + "/jobs/" + JobPathSegment(jobID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status stream: HTTP %d: %s", resp.StatusCode, string(body))
	}
	return readSSEJobStatus(resp.Body, onEvent)
}

// readSSEJobStatus parses SSE events from r, expecting event: job_status and data: JSON JobStatus.
func readSSEJobStatus(r io.Reader, onEvent func(*JobStatus) bool) error {
	scanner := bufio.NewScanner(r)
	var eventType, data string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if eventType == "job_status" && data != "" {
				var s JobStatus
				if err := json.Unmarshal([]byte(data), &s); err != nil {
					eventType = ""
					data = ""
					continue
				}
				if !onEvent(&s) {
					return nil
				}
				if s.Status == "done" || s.Status == "failed" || s.Status == "deleted" {
					return nil
				}
			}
			eventType = ""
			data = ""
			continue
		}
		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			continue
		}
	}
	return scanner.Err()
}

// GetJobResult returns GET /jobs/{job_id}/results.
func (c *Client) GetJobResult(jobID string) ([]byte, int, error) {
	seg := JobPathSegment(jobID)
	return c.Do(http.MethodGet, "/jobs/"+seg+"/results", nil, nil)
}

// DownloadJob writes GET /jobs/{job_id}/download to outPath. Returns written bytes and error.
// outPath must be non-empty; callers should resolve a safe path (e.g. under CWD).
func (c *Client) DownloadJob(jobID, outPath string) (int64, int, error) {
	if outPath == "" {
		return 0, 0, fmt.Errorf("output path is required")
	}
	seg := JobPathSegment(jobID)
	body, status, err := c.Do(http.MethodGet, "/jobs/"+seg+"/download", nil, nil)
	if err != nil {
		return 0, status, err
	}
	if status != http.StatusOK {
		return 0, status, fmt.Errorf("download failed: HTTP %d: %s", status, string(body))
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return 0, status, err
	}
	if err := os.WriteFile(outPath, body, 0644); err != nil {
		return 0, status, err
	}
	return int64(len(body)), status, nil
}

// DeleteJob sends DELETE /jobs/{job_id}.
func (c *Client) DeleteJob(jobID string) ([]byte, int, error) {
	seg := JobPathSegment(jobID)
	return c.Do(http.MethodDelete, "/jobs/"+seg, nil, nil)
}

// GetMe returns GET /me.
func (c *Client) GetMe() ([]byte, int, error) {
	return c.Do(http.MethodGet, "/me", nil, nil)
}

// GetAccountsMe returns GET /me.
func (c *Client) GetAccountsMe() ([]byte, int, error) {
	return c.Do(http.MethodGet, "/me", nil, nil)
}

// GetMyProjects returns GET /myProjects.
func (c *Client) GetMyProjects() ([]byte, int, error) {
	return c.Do(http.MethodGet, "/myProjects", nil, nil)
}

// GetMyUsage returns GET /myUsage.
func (c *Client) GetMyUsage() ([]byte, int, error) {
	return c.Do(http.MethodGet, "/myUsage", nil, nil)
}

// Health returns GET /health (no auth).
func (c *Client) Health() ([]byte, int, error) {
	return c.Do(http.MethodGet, "/health", nil, nil)
}

// PomaArchiveName returns the default filename for a job's POMA archive download.
func PomaArchiveName(jobID string) string {
	return jobID + ".poma"
}

// ParseJob parses the ingest response into Job to get job_id.
func ParseJob(data []byte) (*Job, error) {
	var j Job
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, err
	}
	if j.JobID != "" {
		j.JobID = JobPathSegment(j.JobID)
	}
	return &j, nil
}

// ParseJobStatus parses job status response.
func ParseJobStatus(data []byte) (*JobStatus, error) {
	var s JobStatus
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func sanitizeContentDispositionFilename(name string) string {
	ext := path.Ext(name)
	if name == "" || name == "." || name == ".." {
		return "upload" + ext
	}
	if strings.ContainsAny(name, "\"\\\r\n\x00") {
		return "upload" + ext
	}
	return name
}

// printIngestJobIDOnly writes pretty-printed {"job_id":"..."} to stdout (normalized via ParseJob).
func PrintIngestJobIDOnly(body []byte) error {
	j, err := ParseJob(body)
	if err != nil {
		return fmt.Errorf("parse ingest response: %w", err)
	}
	if j.JobID == "" {
		return fmt.Errorf("ingest response has no job_id")
	}
	out, err := json.MarshalIndent(map[string]string{"job_id": j.JobID}, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
