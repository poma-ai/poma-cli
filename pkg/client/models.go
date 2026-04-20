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

// AccountAPIKeyBody parses GET /me; other JSON fields are ignored.
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

// Orga is the response schema for organisation endpoints.
type Orga struct {
	OrgaID    string `json:"orga_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// OrgaMember is the response schema for organisation member endpoints.
type OrgaMember struct {
	OrgaID    string `json:"orga_id"`
	AccountID string `json:"account_id"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at,omitempty"`
}

// CreateOrgaRequest is the request body for POST /orgas.
type CreateOrgaRequest struct {
	Name string `json:"name"`
}

// UpdateOrgaRequest is the request body for PUT /orgas/{orgaId}.
type UpdateOrgaRequest struct {
	Name string `json:"name"`
}

// AddOrgaMemberRequest is the request body for POST /orgas/{orgaId}/members.
type AddOrgaMemberRequest struct {
	AccountID string `json:"account_id"`
	Role      string `json:"role"`
}

// UpdateOrgaMemberRoleRequest is the request body for PUT /orgas/{orgaId}/members/{accountId}.
type UpdateOrgaMemberRoleRequest struct {
	Role string `json:"role"`
}
