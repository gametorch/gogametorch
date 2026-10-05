package gametorch

import "time"

// ApiKey is an API key. The secret itself is never returned after creation.
type ApiKey struct {
	ID                string      `json:"id"`
	Name              *string     `json:"name"`
	KeyPrefix         string      `json:"key_prefix"`
	ExpiresAt         *time.Time  `json:"expires_at"`
	MaxSpendLimit     *Decimal    `json:"max_spend_limit"`
	SpendResetCadence string      `json:"spend_reset_cadence"`
	KeyScope          ApiKeyScope `json:"key_scope"`
	ProjectID         *string     `json:"project_id"`
	Spend             Decimal     `json:"spend"`
	LifetimeSpend     Decimal     `json:"lifetime_spend"`
	SpendResetAt      *time.Time  `json:"spend_reset_at"`
	CreatedAt         time.Time   `json:"created_at"`
}

// ApiKeyWithSecret is the response from POST /keys: an API key plus its secret,
// returned only once.
type ApiKeyWithSecret struct {
	ApiKey
	// KeyFull is the full key. Store it securely; it cannot be retrieved again.
	KeyFull string `json:"key_full"`
}

// KeysResponse is the response from GET /keys.
type KeysResponse struct {
	Keys []ApiKey `json:"keys"`
}

// CreateApiKeyRequest is the request body for POST /keys.
type CreateApiKeyRequest struct {
	Name              *string            `json:"name,omitempty"`
	ExpiresAt         *time.Time         `json:"expires_at,omitempty"`
	MaxSpendLimit     *Decimal           `json:"max_spend_limit,omitempty"`
	SpendResetCadence *SpendResetCadence `json:"spend_reset_cadence,omitempty"`
	KeyScope          *ApiKeyScope       `json:"key_scope,omitempty"`
	ProjectID         *string            `json:"project_id,omitempty"`
}

// NewCreateApiKeyRequest returns an empty request, which the server treats as
// an admin key.
func NewCreateApiKeyRequest() *CreateApiKeyRequest { return &CreateApiKeyRequest{} }

// AdminCreateApiKeyRequest returns an admin-scoped key request.
func AdminCreateApiKeyRequest() *CreateApiKeyRequest {
	scope := ApiKeyScopeAdmin
	return &CreateApiKeyRequest{KeyScope: &scope}
}

// ProjectWriteCreateApiKeyRequest returns a write-scoped key request bound to
// projectID.
func ProjectWriteCreateApiKeyRequest(projectID string) *CreateApiKeyRequest {
	scope := ApiKeyScopeProjectWrite
	return &CreateApiKeyRequest{KeyScope: &scope, ProjectID: &projectID}
}

// ProjectReadCreateApiKeyRequest returns a read-only key request bound to
// projectID.
func ProjectReadCreateApiKeyRequest(projectID string) *CreateApiKeyRequest {
	scope := ApiKeyScopeProjectRead
	return &CreateApiKeyRequest{KeyScope: &scope, ProjectID: &projectID}
}

// WithName sets the key name.
func (r *CreateApiKeyRequest) WithName(name string) *CreateApiKeyRequest {
	r.Name = &name
	return r
}

// WithExpiresAt sets the expiry time.
func (r *CreateApiKeyRequest) WithExpiresAt(t time.Time) *CreateApiKeyRequest {
	r.ExpiresAt = &t
	return r
}

// WithMaxSpendLimit sets the spend limit for the current cycle.
func (r *CreateApiKeyRequest) WithMaxSpendLimit(limit Decimal) *CreateApiKeyRequest {
	r.MaxSpendLimit = &limit
	return r
}

// WithSpendResetCadence sets the spend reset cadence.
func (r *CreateApiKeyRequest) WithSpendResetCadence(cadence SpendResetCadence) *CreateApiKeyRequest {
	r.SpendResetCadence = &cadence
	return r
}

// WithKeyScope sets the key scope explicitly.
func (r *CreateApiKeyRequest) WithKeyScope(scope ApiKeyScope) *CreateApiKeyRequest {
	r.KeyScope = &scope
	return r
}

// WithProjectID binds the key to a project.
func (r *CreateApiKeyRequest) WithProjectID(projectID string) *CreateApiKeyRequest {
	r.ProjectID = &projectID
	return r
}

// UpdateApiKeyRequest is the request body for PATCH /keys/{id}. Only the fields
// that are set are updated.
type UpdateApiKeyRequest struct {
	Name              *string            `json:"name,omitempty"`
	ExpiresAt         *time.Time         `json:"expires_at,omitempty"`
	MaxSpendLimit     *Decimal           `json:"max_spend_limit,omitempty"`
	SpendResetCadence *SpendResetCadence `json:"spend_reset_cadence,omitempty"`
}

// NewUpdateApiKeyRequest returns an empty request.
func NewUpdateApiKeyRequest() *UpdateApiKeyRequest { return &UpdateApiKeyRequest{} }

// WithName sets the key name.
func (r *UpdateApiKeyRequest) WithName(name string) *UpdateApiKeyRequest {
	r.Name = &name
	return r
}

// WithExpiresAt sets the expiry time.
func (r *UpdateApiKeyRequest) WithExpiresAt(t time.Time) *UpdateApiKeyRequest {
	r.ExpiresAt = &t
	return r
}

// WithMaxSpendLimit sets the spend limit.
func (r *UpdateApiKeyRequest) WithMaxSpendLimit(limit Decimal) *UpdateApiKeyRequest {
	r.MaxSpendLimit = &limit
	return r
}

// WithSpendResetCadence sets the spend reset cadence.
func (r *UpdateApiKeyRequest) WithSpendResetCadence(cadence SpendResetCadence) *UpdateApiKeyRequest {
	r.SpendResetCadence = &cadence
	return r
}
