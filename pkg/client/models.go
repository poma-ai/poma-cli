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

// AccountApiKeyResponse is the response from POST /generateApiKey.
type AccountApiKeyResponse struct {
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
	OrgaID     string `json:"orga_id"`
	Name       string `json:"name"`
	CallerRole string `json:"caller_role,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

// OrgaSearchResult is the response schema for GET /orgas.
type OrgaSearchResult struct {
	Orgas    []Orga `json:"orgas"`
	Total    int    `json:"total"`
	Results  int    `json:"results"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// OrgaMember is the response schema for organisation member endpoints.
type OrgaMember struct {
	OrgaID    string `json:"orga_id"`
	AccountID string `json:"account_id"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at,omitempty"`
}

// OrgaMembersListResponse is the response schema for GET /orgas/{orgaId}/members.
type OrgaMembersListResponse struct {
	Members []OrgaMember `json:"members"`
}

// OrgaInvitation is the response schema for organisation invitation endpoints.
type OrgaInvitation struct {
	ID                  int64  `json:"id"`
	OrgaID              string `json:"orga_id"`
	InviterAccountID    string `json:"inviter_account_id"`
	InviteeEmail        string `json:"invitee_email"`
	Status              string `json:"status"`
	AcceptedByAccountID string `json:"accepted_by_account_id,omitempty"`
	ExpiresAt           string `json:"expires_at,omitempty"`
	ResendCount         int    `json:"resend_count"`
	LastResentAt        string `json:"last_resent_at,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
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
	Email string `json:"email"`
	Role  string `json:"role"`
}

// CreateOrgaInvitationRequest is the request body for POST /orgas/{orgaId}/invitations.
type CreateOrgaInvitationRequest struct {
	Email string `json:"email"`
}

// Project is the response schema for project endpoints.
type Project struct {
	ID          string `json:"id"`
	AccountID   string `json:"account_id"`
	OrgaID      string `json:"orga_id,omitempty"`
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Product     string `json:"product"`
	CreatedAt   string `json:"created_at,omitempty"`
	StorageBase string `json:"storage_base,omitempty"`
}

// CreateProjectRequest is the request body for POST /projects.
type CreateProjectRequest struct {
	Product   string `json:"product"`
	Name      string `json:"name"`
	AccountID string `json:"account_id,omitempty"`
	OrgaID    string `json:"orga_id,omitempty"`
}

// CreateProjectResponse is the response from POST /projects. The api_key is shown only at creation time.
type CreateProjectResponse struct {
	Project Project `json:"project"`
	APIKey  string  `json:"api_key"`
}

// ProjectSearchOptions is the request body for GET /projects/search.
type ProjectSearchOptions struct {
	AccountID string `json:"account_id,omitempty"`
	OrgaID    string `json:"orga_id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Product   string `json:"product,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// ProjectsSearchResult is the response from GET /projects/search.
// Field names match the default Go JSON encoding used by the server.
type ProjectsSearchResult struct {
	Projects []Project `json:"Projects"`
	Total    int       `json:"Total"`
	Results  int       `json:"Results"`
	Page     int       `json:"Page"`
	PageSize int       `json:"PageSize"`
}
