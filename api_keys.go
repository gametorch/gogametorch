package gametorch

import (
	"context"
	"net/http"
)

// ListKeys lists the API keys in the caller's scope. Admin-only for
// organizations.
//
// GET /keys
func (c *Client) ListKeys(ctx context.Context) (*KeysResponse, error) {
	out, err := fetchJSON[KeysResponse](c, ctx, rateTier2, "GET /keys", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "keys",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateKey creates an API key. The full key is returned exactly once.
//
// POST /keys
func (c *Client) CreateKey(ctx context.Context, request *CreateApiKeyRequest) (*ApiKeyWithSecret, error) {
	if request == nil {
		request = NewCreateApiKeyRequest()
	}
	out, err := fetchJSON[ApiKeyWithSecret](c, ctx, rateWrites, "POST /keys", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "keys",
		body:   request,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateKey updates an API key. Only the fields set on request are changed.
//
// PATCH /keys/{id}
func (c *Client) UpdateKey(ctx context.Context, keyID string, request *UpdateApiKeyRequest) (*ApiKey, error) {
	if request == nil {
		request = NewUpdateApiKeyRequest()
	}
	out, err := fetchJSON[ApiKey](c, ctx, rateWrites, "PATCH /keys/{id}", concurrencyNone, requestSpec{
		method: http.MethodPatch,
		path:   "keys/" + keyID,
		body:   request,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteKey revokes an API key.
//
// DELETE /keys/{id}
func (c *Client) DeleteKey(ctx context.Context, keyID string) (*OkResponse, error) {
	return fetchAck(c, ctx, rateWrites, "DELETE /keys/{id}", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "keys/" + keyID,
	})
}
