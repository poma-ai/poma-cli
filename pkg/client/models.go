package client

// AccountRegisterEmailRequest is the request body for POST /registerEmail.
type AccountRegisterEmailRequest struct {
	Email    string `json:"email"`
	Username string `json:"username,omitempty"`
	Company  string `json:"company,omitempty"`
}

// AccountRegisterEmailResponse is the response from POST /registerEmail and POST /verifyEmail (verification returns token).
type AccountRegisterEmailResponse struct {
	Token string `json:"token,omitempty"`
}

// AccountVerifyEmailRequest is the request body for POST /verifyEmail.
type AccountVerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// AccountAPIKeyBody parses GET /accounts/me; other JSON fields are ignored.
type AccountAPIKeyBody struct {
	APIKey string `json:"api_key"`
}

// Job is the response from POST /ingest and POST /ingestEco.
type Job struct {
	JobID string `json:"job_id"`
}

// JobStatus is the response from GET /jobs/{job_id}/status.
// States: "pending", "processing", "done", "failed".
type JobStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
