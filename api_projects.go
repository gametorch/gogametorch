package gametorch

import (
	"context"
	"net/http"
)

type nameBody struct {
	Name string `json:"name"`
}

// ListProjects lists the projects in the caller's scope.
//
// GET /projects
func (c *Client) ListProjects(ctx context.Context) (*ProjectsResponse, error) {
	out, err := fetchJSON[ProjectsResponse](c, ctx, rateTier2, "GET /projects", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "projects",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateProject creates a project owned by the caller's scope.
//
// POST /projects
func (c *Client) CreateProject(ctx context.Context, name string) (*Project, error) {
	out, err := fetchJSON[Project](c, ctx, rateWrites, "POST /projects", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "projects",
		body:   nameBody{Name: name},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// RenameProject renames a project. The slug may change.
//
// PATCH /projects/{slug}
func (c *Client) RenameProject(ctx context.Context, slug, name string) (*Project, error) {
	out, err := fetchJSON[Project](c, ctx, rateWrites, "PATCH /projects/{slug}", concurrencyNone, requestSpec{
		method: http.MethodPatch,
		path:   "projects/" + slug,
		body:   nameBody{Name: name},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteProject permanently deletes a project and its generations and assets.
//
// DELETE /projects/{slug}
func (c *Client) DeleteProject(ctx context.Context, slug string) (*OkResponse, error) {
	return fetchAck(c, ctx, rateWrites, "DELETE /projects/{slug}", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "projects/" + slug,
	})
}
