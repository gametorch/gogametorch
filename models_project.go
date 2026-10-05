package gametorch

import "time"

// Provenance records who or what created a resource. GameTorch records the
// Clerk user and, when the request came through an API key, the key that ran
// the operation. These fields are flattened into the owning resource.
type Provenance struct {
	// UserID is the Clerk user id (or API-key owner) credited for the result.
	UserID string `json:"user_id,omitempty"`
	// Source is the spend/creation source, for example "UI" or "API key".
	Source string `json:"source,omitempty"`
	// APIKeyID is the API key used, when the result came through one.
	APIKeyID *string `json:"api_key_id,omitempty"`
	// KeyName is the API key's name, when available.
	KeyName *string `json:"key_name,omitempty"`
}

// Project is a GameTorch project.
type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

// ProjectsResponse is the response from GET /projects.
type ProjectsResponse struct {
	Projects []Project `json:"projects"`
}
