package gametorch

import (
	"context"
	"net/http"
	"strings"
)

// EnsureUser ensures a mirrored user row exists for the caller.
//
// POST /users/me
func (c *Client) EnsureUser(ctx context.Context) (*OkResponse, error) {
	return fetchAck(c, ctx, rateWrites, "POST /users/me", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "users/me",
	})
}

// Health returns the API health status ("ok" when healthy).
//
// GET /health
func (c *Client) Health(ctx context.Context) (string, error) {
	download, err := fetchDownload(c, ctx, rateUnlimited, "GET /health", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "health",
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(download.Data)), nil
}
