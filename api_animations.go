package gametorch

import (
	"context"
	"net/http"
)

type estimateBody struct {
	AnimationModel string  `json:"animation_model"`
	Duration       *int64  `json:"duration,omitempty"`
	BaseAssetID    *string `json:"base_asset_id,omitempty"`
}

type createAnimationBody struct {
	RequestID      string  `json:"request_id"`
	Prompt         string  `json:"prompt"`
	AnimationModel string  `json:"animation_model"`
	Duration       *int64  `json:"duration,omitempty"`
	BaseAssetID    *string `json:"base_asset_id,omitempty"`
}

type generateFramesBody struct {
	RequestID string `json:"request_id"`
	FPS       int64  `json:"fps"`
}

// EstimateAnimation estimates the credit cost of an animation run without
// starting one. To estimate animating an existing sprite, pass the sprite's
// asset id with WithBaseAssetID.
//
// POST /projects/{project_id}/animation-runs/estimate
func (c *Client) EstimateAnimation(projectID string) *AnimationEstimateBuilder {
	return &AnimationEstimateBuilder{client: c, projectID: projectID}
}

// GenerateAnimation starts building an animation run.
//
// To animate an existing sprite, pass its asset id (from Client.ListSpriteAssets,
// Client.GetAsset or Generation.Assets) with WithBaseAssetID; omit it to
// generate the animation from scratch.
//
// POST /projects/{project_id}/animation-runs
func (c *Client) GenerateAnimation(projectID string) *AnimationRunBuilder {
	return &AnimationRunBuilder{client: c, projectID: projectID}
}

// ListAnimationRuns lists a project's animation runs. Set
// ListParams.BaseAssetID to return only runs based on a given sprite asset.
//
// GET /projects/{project_id}/animation-runs
func (c *Client) ListAnimationRuns(ctx context.Context, projectID string, params *ListParams) (*AnimationsResponse, error) {
	query := listQuery(params)
	if params != nil && params.BaseAssetID != nil {
		query.Set("base_asset_id", *params.BaseAssetID)
	}
	out, err := fetchJSON[AnimationsResponse](c, ctx, rateTier2, "GET /projects/{project_id}/animation-runs", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "projects/" + projectID + "/animation-runs",
		query:  query,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAnimationRun returns one animation run and its frames.
//
// GET /animation-runs/{id}
func (c *Client) GetAnimationRun(ctx context.Context, runID string) (*AnimationRun, error) {
	out, err := fetchJSON[AnimationRun](c, ctx, rateTier2, "GET /animation-runs/{id}", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "animation-runs/" + runID,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AnimationContent returns the animation clip bytes.
//
// GET /animation-runs/{id}/content
func (c *Client) AnimationContent(ctx context.Context, runID string) (*Download, error) {
	return fetchDownload(c, ctx, rateTier2, "GET /animation-runs/{id}/content", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "animation-runs/" + runID + "/content",
	})
}

// ArchiveAnimationRun archives an animation run.
//
// POST /animation-runs/{id}/archive
func (c *Client) ArchiveAnimationRun(ctx context.Context, runID string) (*ArchiveResponse, error) {
	out, err := fetchJSON[ArchiveResponse](c, ctx, rateWrites, "POST /animation-runs/{id}/archive", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "animation-runs/" + runID + "/archive",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UnarchiveAnimationRun restores an archived animation run.
//
// POST /animation-runs/{id}/unarchive
func (c *Client) UnarchiveAnimationRun(ctx context.Context, runID string) (*ArchiveResponse, error) {
	out, err := fetchJSON[ArchiveResponse](c, ctx, rateWrites, "POST /animation-runs/{id}/unarchive", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "animation-runs/" + runID + "/unarchive",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAnimationRun permanently deletes an animation run.
//
// DELETE /animation-runs/{id}
func (c *Client) DeleteAnimationRun(ctx context.Context, runID string) error {
	return fetchOK(c, ctx, rateWrites, "DELETE /animation-runs/{id}", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "animation-runs/" + runID,
	})
}

// GenerateFrames generates individual PNG frames from a finished animation run.
//
// POST /projects/{project_id}/animation-runs/{id}/frames
func (c *Client) GenerateFrames(projectID, runID string) *FrameGenerationBuilder {
	return &FrameGenerationBuilder{client: c, projectID: projectID, runID: runID, fps: 12}
}

// FrameContent returns the PNG bytes of a generated frame by its id.
//
// GET /animation-run-frames/{id}/content
func (c *Client) FrameContent(ctx context.Context, frameID string) (*Download, error) {
	return fetchDownload(c, ctx, rateTier2, "GET /animation-run-frames/{id}/content", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "animation-run-frames/" + frameID + "/content",
	})
}

// FrameContentByNumber returns the PNG bytes of a frame addressed by its 1-based
// number, so a range can be fetched without listing frame ids first.
//
// GET /animation-runs/{id}/frames/{number}/content
func (c *Client) FrameContentByNumber(ctx context.Context, runID string, frameNumber int64) (*Download, error) {
	return fetchDownload(c, ctx, rateTier2, "GET /animation-runs/{id}/frames/{number}/content", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "animation-runs/" + runID + "/frames/" + int64String(frameNumber) + "/content",
	})
}

// StreamAnimationRuns streams a project's animation runs page by page, honoring
// the same filters as Client.ListAnimationRuns (including BaseAssetID).
func (c *Client) StreamAnimationRuns(projectID string, params *ListParams) *Paginator[AnimationRun] {
	base := &ListParams{}
	if params != nil {
		base.IncludeArchived = params.IncludeArchived
		base.BaseAssetID = params.BaseAssetID
	}
	return newPaginator(func(ctx context.Context, cursor *string) (Page[AnimationRun], error) {
		pageParams := &ListParams{
			Before:          cursor,
			IncludeArchived: base.IncludeArchived,
			BaseAssetID:     base.BaseAssetID,
		}
		response, err := c.ListAnimationRuns(ctx, projectID, pageParams)
		if err != nil {
			return Page[AnimationRun]{}, err
		}
		return Page[AnimationRun]{
			Items:      response.Animations,
			NextCursor: response.NextCursor,
			Total:      response.Total,
		}, nil
	})
}

// AnimationEstimateBuilder builds an animation cost estimate.
type AnimationEstimateBuilder struct {
	client         *Client
	projectID      string
	animationModel string
	duration       *int64
	baseAssetID    *string
}

// WithAnimationModel sets the animation model (ash, birch or cedar); required.
func (b *AnimationEstimateBuilder) WithAnimationModel(animationModel string) *AnimationEstimateBuilder {
	b.animationModel = animationModel
	return b
}

// WithDuration sets the desired duration in seconds.
func (b *AnimationEstimateBuilder) WithDuration(duration int64) *AnimationEstimateBuilder {
	b.duration = &duration
	return b
}

// WithBaseAssetID bases the estimate on an existing sprite asset. Pass a sprite
// asset id from Client.ListSpriteAssets, Client.GetAsset or Generation.Assets.
func (b *AnimationEstimateBuilder) WithBaseAssetID(baseAssetID string) *AnimationEstimateBuilder {
	b.baseAssetID = &baseAssetID
	return b
}

// Send sends the estimate request.
func (b *AnimationEstimateBuilder) Send(ctx context.Context) (*AnimationEstimate, error) {
	if b.animationModel == "" {
		return nil, newConfigError("animation estimate requires an animation_model")
	}
	body := estimateBody{
		AnimationModel: b.animationModel,
		Duration:       b.duration,
		BaseAssetID:    b.baseAssetID,
	}
	out, err := fetchJSON[AnimationEstimate](b.client, ctx, rateTier2, "POST /projects/{project_id}/animation-runs/estimate", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "projects/" + b.projectID + "/animation-runs/estimate",
		body:   body,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AnimationRunBuilder builds an animation run.
type AnimationRunBuilder struct {
	client         *Client
	projectID      string
	prompt         string
	animationModel string
	duration       *int64
	baseAssetID    *string
	requestID      *string
}

// WithPrompt sets the generation prompt (required). When a base asset is set,
// the prompt describes the motion to apply to that sprite.
func (b *AnimationRunBuilder) WithPrompt(prompt string) *AnimationRunBuilder {
	b.prompt = prompt
	return b
}

// WithAnimationModel sets the animation model (ash, birch or cedar); required.
func (b *AnimationRunBuilder) WithAnimationModel(animationModel string) *AnimationRunBuilder {
	b.animationModel = animationModel
	return b
}

// WithDuration sets the desired duration in seconds.
func (b *AnimationRunBuilder) WithDuration(duration int64) *AnimationRunBuilder {
	b.duration = &duration
	return b
}

// WithBaseAssetID animates an existing sprite asset instead of generating from
// scratch. Pass a sprite asset id from Client.ListSpriteAssets, Client.GetAsset
// or Generation.Assets; the resulting AnimationRun reports it as BaseAssetID.
func (b *AnimationRunBuilder) WithBaseAssetID(baseAssetID string) *AnimationRunBuilder {
	b.baseAssetID = &baseAssetID
	return b
}

// WithRequestID overrides the idempotency key. Defaults to a fresh UUID v4.
func (b *AnimationRunBuilder) WithRequestID(requestID string) *AnimationRunBuilder {
	b.requestID = &requestID
	return b
}

// Send sends the animation run request.
func (b *AnimationRunBuilder) Send(ctx context.Context) (*JobCreated, error) {
	if b.prompt == "" {
		return nil, newConfigError("animation run requires a prompt")
	}
	if b.animationModel == "" {
		return nil, newConfigError("animation run requires an animation_model")
	}
	requestID := newUUIDv4()
	if b.requestID != nil {
		requestID = *b.requestID
	}
	body := createAnimationBody{
		RequestID:      requestID,
		Prompt:         b.prompt,
		AnimationModel: b.animationModel,
		Duration:       b.duration,
		BaseAssetID:    b.baseAssetID,
	}
	out, err := fetchJSON[JobCreated](b.client, ctx, rateTier1, "POST /projects/{project_id}/animation-runs", concurrencyHold, requestSpec{
		method: http.MethodPost,
		path:   "projects/" + b.projectID + "/animation-runs",
		body:   body,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// FrameGenerationBuilder builds animation frame generation.
type FrameGenerationBuilder struct {
	client    *Client
	projectID string
	runID     string
	fps       int64
	requestID *string
}

// WithFPS sets the sampling rate (1-30 fps). Defaults to 12.
func (b *FrameGenerationBuilder) WithFPS(fps int64) *FrameGenerationBuilder {
	b.fps = fps
	return b
}

// WithRequestID overrides the idempotency key. Defaults to a fresh UUID v4.
func (b *FrameGenerationBuilder) WithRequestID(requestID string) *FrameGenerationBuilder {
	b.requestID = &requestID
	return b
}

// Send sends the frame-generation request.
func (b *FrameGenerationBuilder) Send(ctx context.Context) (*JobCreated, error) {
	requestID := newUUIDv4()
	if b.requestID != nil {
		requestID = *b.requestID
	}
	body := generateFramesBody{RequestID: requestID, FPS: b.fps}
	out, err := fetchJSON[JobCreated](b.client, ctx, rateTier1, "POST /projects/{project_id}/animation-runs/{id}/frames", concurrencyFrame, requestSpec{
		method: http.MethodPost,
		path:   "projects/" + b.projectID + "/animation-runs/" + b.runID + "/frames",
		body:   body,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
