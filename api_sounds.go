package gametorch

import (
	"context"
	"net/http"
	"net/url"
)

type createSoundBody struct {
	RequestID      string  `json:"request_id"`
	Prompt         string  `json:"prompt"`
	SoundModel     string  `json:"sound_model"`
	ResponseFormat *string `json:"response_format,omitempty"`
}

type updateSoundBody struct {
	Name *string `json:"name"`
}

// GenerateSound starts building a sound generation.
func (c *Client) GenerateSound(projectID string) *SoundGenerationBuilder {
	return &SoundGenerationBuilder{client: c, projectID: projectID}
}

// ListSoundGenerations lists a project's sound generations.
//
// GET /projects/{project_id}/sound-generations
func (c *Client) ListSoundGenerations(ctx context.Context, projectID string, params *ListParams) (*SoundGenerationsResponse, error) {
	out, err := fetchJSON[SoundGenerationsResponse](c, ctx, rateTier2, "GET /projects/{project_id}/sound-generations", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "projects/" + projectID + "/sound-generations",
		query:  listQuery(params),
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSoundGeneration returns one sound generation and its assets.
//
// GET /sound-generations/{id}
func (c *Client) GetSoundGeneration(ctx context.Context, generationID string, includeArchived bool) (*SoundGeneration, error) {
	query := url.Values{}
	if includeArchived {
		query.Set("include_archived", "true")
	}
	out, err := fetchJSON[SoundGeneration](c, ctx, rateTier2, "GET /sound-generations/{id}", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "sound-generations/" + generationID,
		query:  query,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// SoundAssetContent returns the audio bytes of a sound asset.
//
// GET /sound-assets/{id}/content
func (c *Client) SoundAssetContent(ctx context.Context, assetID string) (*Download, error) {
	return fetchDownload(c, ctx, rateTier2, "GET /sound-assets/{id}/content", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "sound-assets/" + assetID + "/content",
	})
}

// RenameSoundAsset renames a sound asset. Pass nil to clear the name.
//
// PATCH /sound-assets/{id}
func (c *Client) RenameSoundAsset(ctx context.Context, assetID string, name *string) (*AssetNameResponse, error) {
	out, err := fetchJSON[AssetNameResponse](c, ctx, rateWrites, "PATCH /sound-assets/{id}", concurrencyNone, requestSpec{
		method: http.MethodPatch,
		path:   "sound-assets/" + assetID,
		body:   updateSoundBody{Name: name},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// PutSoundAssetMetadata replaces a sound asset's metadata.
//
// PUT /sound-assets/{id}/metadata
func (c *Client) PutSoundAssetMetadata(ctx context.Context, assetID string, metadata map[string]string) (*AssetMetadataResponse, error) {
	if metadata == nil {
		metadata = map[string]string{}
	}
	out, err := fetchJSON[AssetMetadataResponse](c, ctx, rateWrites, "PUT /sound-assets/{id}/metadata", concurrencyNone, requestSpec{
		method: http.MethodPut,
		path:   "sound-assets/" + assetID + "/metadata",
		body:   metadataBody{Metadata: metadata},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ArchiveSoundAsset archives a sound asset.
//
// POST /sound-assets/{id}/archive
func (c *Client) ArchiveSoundAsset(ctx context.Context, assetID string) (*ArchiveResponse, error) {
	out, err := fetchJSON[ArchiveResponse](c, ctx, rateWrites, "POST /sound-assets/{id}/archive", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "sound-assets/" + assetID + "/archive",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UnarchiveSoundAsset restores an archived sound asset.
//
// POST /sound-assets/{id}/unarchive
func (c *Client) UnarchiveSoundAsset(ctx context.Context, assetID string) (*ArchiveResponse, error) {
	out, err := fetchJSON[ArchiveResponse](c, ctx, rateWrites, "POST /sound-assets/{id}/unarchive", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "sound-assets/" + assetID + "/unarchive",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSoundAsset permanently deletes a sound asset.
//
// DELETE /sound-assets/{id}
func (c *Client) DeleteSoundAsset(ctx context.Context, assetID string) error {
	return fetchOK(c, ctx, rateWrites, "DELETE /sound-assets/{id}", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "sound-assets/" + assetID,
	})
}

// ArchiveSoundGeneration archives a whole sound generation.
//
// POST /sound-generations/{id}/archive
func (c *Client) ArchiveSoundGeneration(ctx context.Context, generationID string) (*ArchiveResponse, error) {
	out, err := fetchJSON[ArchiveResponse](c, ctx, rateWrites, "POST /sound-generations/{id}/archive", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "sound-generations/" + generationID + "/archive",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UnarchiveSoundGeneration restores an archived sound generation.
//
// POST /sound-generations/{id}/unarchive
func (c *Client) UnarchiveSoundGeneration(ctx context.Context, generationID string) (*ArchiveResponse, error) {
	out, err := fetchJSON[ArchiveResponse](c, ctx, rateWrites, "POST /sound-generations/{id}/unarchive", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "sound-generations/" + generationID + "/unarchive",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// StreamSoundGenerations streams a project's sound generations page by page.
func (c *Client) StreamSoundGenerations(projectID string, includeArchived bool) *Paginator[SoundGeneration] {
	return newPaginator(func(ctx context.Context, cursor *string) (Page[SoundGeneration], error) {
		params := &ListParams{Before: cursor, IncludeArchived: includeArchived}
		response, err := c.ListSoundGenerations(ctx, projectID, params)
		if err != nil {
			return Page[SoundGeneration]{}, err
		}
		return Page[SoundGeneration]{
			Items:      response.SoundGenerations,
			NextCursor: response.NextCursor,
			Total:      response.Total,
		}, nil
	})
}

// SoundGenerationBuilder builds a sound generation. Prompt and SoundModel are
// required.
type SoundGenerationBuilder struct {
	client         *Client
	projectID      string
	prompt         string
	soundModel     string
	responseFormat *string
	requestID      *string
}

// WithPrompt sets the generation prompt (required).
func (b *SoundGenerationBuilder) WithPrompt(prompt string) *SoundGenerationBuilder {
	b.prompt = prompt
	return b
}

// WithSoundModel sets the sound model id (required). See Client.SoundModels.
func (b *SoundGenerationBuilder) WithSoundModel(soundModel string) *SoundGenerationBuilder {
	b.soundModel = soundModel
	return b
}

// WithResponseFormat sets the output format, for example "mp3" or "pcm".
func (b *SoundGenerationBuilder) WithResponseFormat(responseFormat string) *SoundGenerationBuilder {
	b.responseFormat = &responseFormat
	return b
}

// WithRequestID overrides the idempotency key. Defaults to a fresh UUID v4.
func (b *SoundGenerationBuilder) WithRequestID(requestID string) *SoundGenerationBuilder {
	b.requestID = &requestID
	return b
}

// Send sends the sound generation request.
func (b *SoundGenerationBuilder) Send(ctx context.Context) (*JobCreated, error) {
	if b.prompt == "" {
		return nil, newConfigError("sound generation requires a prompt")
	}
	if b.soundModel == "" {
		return nil, newConfigError("sound generation requires a sound_model")
	}
	requestID := newUUIDv4()
	if b.requestID != nil {
		requestID = *b.requestID
	}
	body := createSoundBody{
		RequestID:      requestID,
		Prompt:         b.prompt,
		SoundModel:     b.soundModel,
		ResponseFormat: b.responseFormat,
	}
	out, err := fetchJSON[JobCreated](b.client, ctx, rateTier1, "POST /projects/{project_id}/sound-generations", concurrencyHold, requestSpec{
		method: http.MethodPost,
		path:   "projects/" + b.projectID + "/sound-generations",
		body:   body,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
