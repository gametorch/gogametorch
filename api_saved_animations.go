package gametorch

import (
	"context"
	"net/http"
)

type createSavedAnimationBody struct {
	GenerationID string  `json:"generation_id"`
	StartFrame   int64   `json:"start_frame"`
	EndFrame     int64   `json:"end_frame"`
	Name         *string `json:"name,omitempty"`
}

type updateSavedAnimationBody struct {
	Name *string `json:"name"`
}

// ListSavedAnimations lists a project's saved animations.
//
// GET /projects/{project_id}/saved-animations
func (c *Client) ListSavedAnimations(ctx context.Context, projectID string, params *ListParams) (*SavedAnimationsResponse, error) {
	out, err := fetchJSON[SavedAnimationsResponse](c, ctx, rateTier2, "GET /projects/{project_id}/saved-animations", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "projects/" + projectID + "/saved-animations",
		query:  listQuery(params),
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// SaveAnimation starts building a saved animation from a range of an animation
// run.
//
// POST /projects/{project_id}/saved-animations
func (c *Client) SaveAnimation(projectID string) *SaveAnimationBuilder {
	return &SaveAnimationBuilder{client: c, projectID: projectID}
}

// GetSavedAnimation returns one saved animation.
//
// GET /saved-animations/{id}
func (c *Client) GetSavedAnimation(ctx context.Context, id string) (*SavedAnimation, error) {
	out, err := fetchJSON[SavedAnimation](c, ctx, rateTier2, "GET /saved-animations/{id}", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "saved-animations/" + id,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// RenameSavedAnimation renames a saved animation. Pass nil to clear the name.
//
// PATCH /saved-animations/{id}
func (c *Client) RenameSavedAnimation(ctx context.Context, id string, name *string) (*AssetNameResponse, error) {
	out, err := fetchJSON[AssetNameResponse](c, ctx, rateWrites, "PATCH /saved-animations/{id}", concurrencyNone, requestSpec{
		method: http.MethodPatch,
		path:   "saved-animations/" + id,
		body:   updateSavedAnimationBody{Name: name},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// PutSavedAnimationMetadata replaces a saved animation's metadata.
//
// PUT /saved-animations/{id}/metadata
func (c *Client) PutSavedAnimationMetadata(ctx context.Context, id string, metadata map[string]string) (*AssetMetadataResponse, error) {
	if metadata == nil {
		metadata = map[string]string{}
	}
	out, err := fetchJSON[AssetMetadataResponse](c, ctx, rateWrites, "PUT /saved-animations/{id}/metadata", concurrencyNone, requestSpec{
		method: http.MethodPut,
		path:   "saved-animations/" + id + "/metadata",
		body:   metadataBody{Metadata: metadata},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ArchiveSavedAnimation archives a saved animation.
//
// POST /saved-animations/{id}/archive
func (c *Client) ArchiveSavedAnimation(ctx context.Context, id string) (*ArchiveResponse, error) {
	out, err := fetchJSON[ArchiveResponse](c, ctx, rateWrites, "POST /saved-animations/{id}/archive", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "saved-animations/" + id + "/archive",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UnarchiveSavedAnimation restores an archived saved animation.
//
// POST /saved-animations/{id}/unarchive
func (c *Client) UnarchiveSavedAnimation(ctx context.Context, id string) (*ArchiveResponse, error) {
	out, err := fetchJSON[ArchiveResponse](c, ctx, rateWrites, "POST /saved-animations/{id}/unarchive", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "saved-animations/" + id + "/unarchive",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSavedAnimation permanently deletes a saved animation.
//
// DELETE /saved-animations/{id}
func (c *Client) DeleteSavedAnimation(ctx context.Context, id string) error {
	return fetchOK(c, ctx, rateWrites, "DELETE /saved-animations/{id}", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "saved-animations/" + id,
	})
}

// StreamSavedAnimations streams a project's saved animations page by page.
func (c *Client) StreamSavedAnimations(projectID string, includeArchived bool) *Paginator[SavedAnimation] {
	return newPaginator(func(ctx context.Context, cursor *string) (Page[SavedAnimation], error) {
		params := &ListParams{Before: cursor, IncludeArchived: includeArchived}
		response, err := c.ListSavedAnimations(ctx, projectID, params)
		if err != nil {
			return Page[SavedAnimation]{}, err
		}
		return Page[SavedAnimation]{
			Items:      response.SavedAnimations,
			NextCursor: response.NextCursor,
			Total:      response.Total,
		}, nil
	})
}

// SaveAnimationBuilder builds a saved animation.
type SaveAnimationBuilder struct {
	client       *Client
	projectID    string
	generationID string
	startFrame   *int64
	endFrame     *int64
	name         *string
}

// WithGenerationID sets the source animation run id (required).
func (b *SaveAnimationBuilder) WithGenerationID(generationID string) *SaveAnimationBuilder {
	b.generationID = generationID
	return b
}

// WithRange sets the inclusive frame range (required).
func (b *SaveAnimationBuilder) WithRange(startFrame, endFrame int64) *SaveAnimationBuilder {
	b.startFrame = &startFrame
	b.endFrame = &endFrame
	return b
}

// WithName sets an optional name for the saved animation.
func (b *SaveAnimationBuilder) WithName(name string) *SaveAnimationBuilder {
	b.name = &name
	return b
}

// Send creates the saved animation.
func (b *SaveAnimationBuilder) Send(ctx context.Context) (*SavedAnimation, error) {
	if b.generationID == "" {
		return nil, newConfigError("saved animation requires a generation_id")
	}
	if b.startFrame == nil {
		return nil, newConfigError("saved animation requires a start_frame")
	}
	if b.endFrame == nil {
		return nil, newConfigError("saved animation requires an end_frame")
	}
	body := createSavedAnimationBody{
		GenerationID: b.generationID,
		StartFrame:   *b.startFrame,
		EndFrame:     *b.endFrame,
		Name:         b.name,
	}
	out, err := fetchJSON[SavedAnimation](b.client, ctx, rateWrites, "POST /projects/{project_id}/saved-animations", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "projects/" + b.projectID + "/saved-animations",
		body:   body,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
