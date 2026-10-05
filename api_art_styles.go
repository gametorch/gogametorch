package gametorch

import (
	"context"
	"net/http"
)

type createArtStyleBody struct {
	Name string `json:"name"`
}

// ListArtStyles lists a project's reusable art styles.
//
// GET /projects/{project_id}/art-styles
func (c *Client) ListArtStyles(ctx context.Context, projectID string) (*ArtStylesResponse, error) {
	out, err := fetchJSON[ArtStylesResponse](c, ctx, rateTier2, "GET /projects/{project_id}/art-styles", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "projects/" + projectID + "/art-styles",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateArtStyle creates an art style.
//
// POST /projects/{project_id}/art-styles
func (c *Client) CreateArtStyle(ctx context.Context, projectID, name string) (*ArtStyle, error) {
	out, err := fetchJSON[ArtStyle](c, ctx, rateWrites, "POST /projects/{project_id}/art-styles", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "projects/" + projectID + "/art-styles",
		body:   createArtStyleBody{Name: name},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GenerateArtStyle returns a fresh suggested art style for the user to review.
//
// POST /projects/{project_id}/art-styles/generate
func (c *Client) GenerateArtStyle(ctx context.Context, projectID string) (*ArtStyleSuggestion, error) {
	out, err := fetchJSON[ArtStyleSuggestion](c, ctx, rateTier2, "POST /projects/{project_id}/art-styles/generate", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "projects/" + projectID + "/art-styles/generate",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteArtStyle deletes an art style.
//
// DELETE /art-styles/{art_style_id}
func (c *Client) DeleteArtStyle(ctx context.Context, artStyleID string) (*OkResponse, error) {
	return fetchAck(c, ctx, rateWrites, "DELETE /art-styles/{art_style_id}", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "art-styles/" + artStyleID,
	})
}
