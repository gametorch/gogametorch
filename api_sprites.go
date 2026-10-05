package gametorch

import (
	"context"
	"net/http"
	"net/url"
)

type createGenerationBody struct {
	RequestID   string     `json:"request_id"`
	Prompt      string     `json:"prompt"`
	Mode        SpriteMode `json:"mode"`
	ImageModel  string     `json:"image_model"`
	TextModel   *string    `json:"text_model,omitempty"`
	Quality     *string    `json:"quality,omitempty"`
	Resolution  *string    `json:"resolution,omitempty"`
	BaseAssetID *string    `json:"base_asset_id,omitempty"`
}

type updateAssetBody struct {
	Name *string `json:"name"`
}

type metadataBody struct {
	Metadata map[string]string `json:"metadata"`
}

// GenerateSprite starts building a sprite generation.
func (c *Client) GenerateSprite(projectID string) *SpriteGenerationBuilder {
	return &SpriteGenerationBuilder{
		client:    c,
		projectID: projectID,
		mode:      SpriteModeSingle,
	}
}

// ListGenerations lists a project's generations, including their assets.
//
// GET /projects/{project_id}/generations
func (c *Client) ListGenerations(ctx context.Context, projectID string, params *ListParams) (*GenerationsResponse, error) {
	out, err := fetchJSON[GenerationsResponse](c, ctx, rateTier2, "GET /projects/{project_id}/generations", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "projects/" + projectID + "/generations",
		query:  listQuery(params),
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetGeneration returns one generation and its results.
//
// GET /generations/{id}
func (c *Client) GetGeneration(ctx context.Context, generationID string, includeArchived bool) (*Generation, error) {
	query := url.Values{}
	if includeArchived {
		query.Set("include_archived", "true")
	}
	out, err := fetchJSON[Generation](c, ctx, rateTier2, "GET /generations/{id}", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "generations/" + generationID,
		query:  query,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSpriteAssets searches a project's sprite assets by name or label. Pass a
// nil query to list everything.
//
// GET /projects/{project_id}/sprite-assets
func (c *Client) ListSpriteAssets(ctx context.Context, projectID string, query *string) (*SpriteAssetsResponse, error) {
	values := url.Values{}
	if query != nil {
		values.Set("q", *query)
	}
	out, err := fetchJSON[SpriteAssetsResponse](c, ctx, rateTier2, "GET /projects/{project_id}/sprite-assets", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "projects/" + projectID + "/sprite-assets",
		query:  values,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAsset returns one sprite asset.
//
// GET /assets/{id}
func (c *Client) GetAsset(ctx context.Context, assetID string) (*Asset, error) {
	out, err := fetchJSON[Asset](c, ctx, rateTier2, "GET /assets/{id}", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "assets/" + assetID,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AssetContent returns the trimmed PNG bytes of a sprite asset.
//
// GET /assets/{id}/content
func (c *Client) AssetContent(ctx context.Context, assetID string) (*Download, error) {
	return fetchDownload(c, ctx, rateTier2, "GET /assets/{id}/content", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "assets/" + assetID + "/content",
	})
}

// AssetOriginal returns the uncropped original PNG bytes of a sprite asset.
//
// GET /assets/{id}/original
func (c *Client) AssetOriginal(ctx context.Context, assetID string) (*Download, error) {
	return fetchDownload(c, ctx, rateTier2, "GET /assets/{id}/original", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "assets/" + assetID + "/original",
	})
}

// RenameAsset renames a sprite asset. Pass nil to clear the name.
//
// PATCH /assets/{id}
func (c *Client) RenameAsset(ctx context.Context, assetID string, name *string) (*AssetNameResponse, error) {
	out, err := fetchJSON[AssetNameResponse](c, ctx, rateWrites, "PATCH /assets/{id}", concurrencyNone, requestSpec{
		method: http.MethodPatch,
		path:   "assets/" + assetID,
		body:   updateAssetBody{Name: name},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// PutAssetMetadata replaces a sprite asset's metadata.
//
// PUT /assets/{id}/metadata
func (c *Client) PutAssetMetadata(ctx context.Context, assetID string, metadata map[string]string) (*AssetMetadataResponse, error) {
	if metadata == nil {
		metadata = map[string]string{}
	}
	out, err := fetchJSON[AssetMetadataResponse](c, ctx, rateWrites, "PUT /assets/{id}/metadata", concurrencyNone, requestSpec{
		method: http.MethodPut,
		path:   "assets/" + assetID + "/metadata",
		body:   metadataBody{Metadata: metadata},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ArchiveAsset archives a sprite asset, hiding it from default lists.
//
// POST /assets/{id}/archive
func (c *Client) ArchiveAsset(ctx context.Context, assetID string) (*ArchiveResponse, error) {
	out, err := fetchJSON[ArchiveResponse](c, ctx, rateWrites, "POST /assets/{id}/archive", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "assets/" + assetID + "/archive",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UnarchiveAsset restores an archived sprite asset.
//
// POST /assets/{id}/unarchive
func (c *Client) UnarchiveAsset(ctx context.Context, assetID string) (*ArchiveResponse, error) {
	out, err := fetchJSON[ArchiveResponse](c, ctx, rateWrites, "POST /assets/{id}/unarchive", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "assets/" + assetID + "/unarchive",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAsset permanently deletes a sprite asset and its original.
//
// DELETE /assets/{id}
func (c *Client) DeleteAsset(ctx context.Context, assetID string) error {
	return fetchOK(c, ctx, rateWrites, "DELETE /assets/{id}", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "assets/" + assetID,
	})
}

// ArchiveGeneration archives a whole generation.
//
// POST /generations/{id}/archive
func (c *Client) ArchiveGeneration(ctx context.Context, generationID string) (*ArchiveResponse, error) {
	out, err := fetchJSON[ArchiveResponse](c, ctx, rateWrites, "POST /generations/{id}/archive", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "generations/" + generationID + "/archive",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UnarchiveGeneration restores an archived generation.
//
// POST /generations/{id}/unarchive
func (c *Client) UnarchiveGeneration(ctx context.Context, generationID string) (*ArchiveResponse, error) {
	out, err := fetchJSON[ArchiveResponse](c, ctx, rateWrites, "POST /generations/{id}/unarchive", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "generations/" + generationID + "/unarchive",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteGeneration permanently deletes a generation.
//
// DELETE /generations/{id}
func (c *Client) DeleteGeneration(ctx context.Context, generationID string) error {
	return fetchOK(c, ctx, rateWrites, "DELETE /generations/{id}", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "generations/" + generationID,
	})
}

// StreamGenerations streams a project's generations page by page.
func (c *Client) StreamGenerations(projectID string, includeArchived bool) *Paginator[Generation] {
	return newPaginator(func(ctx context.Context, cursor *string) (Page[Generation], error) {
		params := &ListParams{Before: cursor, IncludeArchived: includeArchived}
		response, err := c.ListGenerations(ctx, projectID, params)
		if err != nil {
			return Page[Generation]{}, err
		}
		return Page[Generation]{
			Items:      response.Generations,
			NextCursor: response.NextCursor,
			Total:      response.Total,
		}, nil
	})
}

// SpriteGenerationBuilder builds a sprite generation. Prompt and ImageModel are
// required.
type SpriteGenerationBuilder struct {
	client      *Client
	projectID   string
	prompt      string
	mode        SpriteMode
	imageModel  string
	textModel   *string
	quality     *string
	resolution  *string
	baseAssetID *string
	requestID   *string
}

// WithPrompt sets the generation prompt (required).
func (b *SpriteGenerationBuilder) WithPrompt(prompt string) *SpriteGenerationBuilder {
	b.prompt = prompt
	return b
}

// WithMode sets the generation mode. Defaults to SpriteModeSingle.
func (b *SpriteGenerationBuilder) WithMode(mode SpriteMode) *SpriteGenerationBuilder {
	b.mode = mode
	return b
}

// WithImageModel sets the image model id (required). See Client.SpriteModels.
func (b *SpriteGenerationBuilder) WithImageModel(imageModel string) *SpriteGenerationBuilder {
	b.imageModel = imageModel
	return b
}

// WithTextModel sets the prompt-enhancement text model. Use "none" to disable
// enhancement.
func (b *SpriteGenerationBuilder) WithTextModel(textModel string) *SpriteGenerationBuilder {
	b.textModel = &textModel
	return b
}

// WithNoTextModel disables prompt enhancement.
func (b *SpriteGenerationBuilder) WithNoTextModel() *SpriteGenerationBuilder {
	none := "none"
	b.textModel = &none
	return b
}

// WithQuality sets the quality setting.
func (b *SpriteGenerationBuilder) WithQuality(quality string) *SpriteGenerationBuilder {
	b.quality = &quality
	return b
}

// WithResolution sets the resolution setting.
func (b *SpriteGenerationBuilder) WithResolution(resolution string) *SpriteGenerationBuilder {
	b.resolution = &resolution
	return b
}

// WithBaseAssetID edits an existing individual image result.
func (b *SpriteGenerationBuilder) WithBaseAssetID(baseAssetID string) *SpriteGenerationBuilder {
	b.baseAssetID = &baseAssetID
	return b
}

// WithRequestID overrides the idempotency key. By default a fresh UUID v4 is
// generated for every call. Reusing a key with the same body returns the
// existing generation; reusing it with a different body returns 409.
func (b *SpriteGenerationBuilder) WithRequestID(requestID string) *SpriteGenerationBuilder {
	b.requestID = &requestID
	return b
}

// Send sends the generation request.
func (b *SpriteGenerationBuilder) Send(ctx context.Context) (*JobCreated, error) {
	if b.prompt == "" {
		return nil, newConfigError("sprite generation requires a prompt")
	}
	if b.imageModel == "" {
		return nil, newConfigError("sprite generation requires an image_model")
	}
	requestID := newUUIDv4()
	if b.requestID != nil {
		requestID = *b.requestID
	}
	body := createGenerationBody{
		RequestID:   requestID,
		Prompt:      b.prompt,
		Mode:        b.mode,
		ImageModel:  b.imageModel,
		TextModel:   b.textModel,
		Quality:     b.quality,
		Resolution:  b.resolution,
		BaseAssetID: b.baseAssetID,
	}
	out, err := fetchJSON[JobCreated](b.client, ctx, rateTier1, "POST /projects/{project_id}/generations", concurrencyHold, requestSpec{
		method: http.MethodPost,
		path:   "projects/" + b.projectID + "/generations",
		body:   body,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
