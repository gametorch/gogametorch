package gametorch

import (
	"encoding/json"
	"time"
)

// AnimationRun is an animation run and its results.
type AnimationRun struct {
	ID              string          `json:"id"`
	ProjectID       string          `json:"project_id"`
	Prompt          string          `json:"prompt"`
	AnimationModel  *string         `json:"animation_model"`
	Duration        *int64          `json:"duration"`
	Animation       *AnimationAsset `json:"animation"`
	Status          string          `json:"status"`
	CreditsConsumed Decimal         `json:"credits_consumed"`
	ReservedCredits Decimal         `json:"reserved_credits"`
	AssetsDelivered BoolOrInt       `json:"assets_delivered"`
	// BaseAssetID is the id of the sprite asset this run was generated from
	// (passed to AnimationRunBuilder.WithBaseAssetID), or nil when generated
	// from scratch.
	BaseAssetID *string          `json:"base_asset_id"`
	Error       *string          `json:"error"`
	CreatedAt   time.Time        `json:"created_at"`
	CompletedAt *time.Time       `json:"completed_at"`
	ArchivedAt  *time.Time       `json:"archived_at"`
	FrameRuns   []FrameRun       `json:"frame_runs"`
	Frames      []AnimationFrame `json:"frames"`
	Provenance
}

// AnimationAsset is metadata about the underlying animation clip.
type AnimationAsset struct {
	ID              string    `json:"id"`
	DurationSeconds int64     `json:"duration_seconds"`
	Resolution      string    `json:"resolution"`
	CreatedAt       time.Time `json:"created_at"`
}

// FrameRun is a frame-generation run that sampled an animation into PNG frames.
type FrameRun struct {
	ID              string    `json:"id"`
	Status          string    `json:"status"`
	FPS             int64     `json:"fps"`
	FrameCount      int64     `json:"frame_count"`
	Warming         bool      `json:"warming"`
	CreditsConsumed Decimal   `json:"credits_consumed"`
	ReservedCredits Decimal   `json:"reserved_credits"`
	Error           *string   `json:"error"`
	CreatedAt       time.Time `json:"created_at"`
}

// AnimationFrame is an individual generated animation frame.
type AnimationFrame struct {
	ID           string    `json:"id"`
	FrameNumber  int64     `json:"frame_number"`
	GenerationID *string   `json:"generation_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// AnimationsResponse is the response from GET /projects/{project_id}/animation-runs.
type AnimationsResponse struct {
	Animations []AnimationRun `json:"animations"`
	NextCursor *string        `json:"next_cursor"`
	Total      int64          `json:"total"`
}

// AnimationEstimate is the response from POST
// /projects/{project_id}/animation-runs/estimate.
type AnimationEstimate struct {
	AnimationModel  string  `json:"animation_model"`
	Duration        int64   `json:"duration"`
	Resolution      string  `json:"resolution"`
	Credits         Decimal `json:"credits"`
	USD             Decimal `json:"usd"`
	ReservedCredits Decimal `json:"reserved_credits"`
}

// ExportPlan is the response from POST /animation-runs/{id}/export-plan.
type ExportPlan struct {
	StartFrame   int32            `json:"start_frame"`
	EndFrame     int32            `json:"end_frame"`
	Reference    *ExportReference `json:"reference"`
	FirstFrame   *ExportFrame     `json:"first_frame"`
	Frames       []ExportFrame    `json:"frames"`
	MaxWidth     int64            `json:"max_width"`
	MaxHeight    int64            `json:"max_height"`
	Scale        float64          `json:"scale"`
	ScaleWidth   float64          `json:"scale_width"`
	ScaleHeight  float64          `json:"scale_height"`
	CanvasWidth  int64            `json:"canvas_width"`
	CanvasHeight int64            `json:"canvas_height"`
	FrameCount   int64            `json:"frame_count"`
}

// ExportReference is the reference frame used to size an export canvas.
type ExportReference struct {
	Bounds      []int64 `json:"bounds"`
	ImageWidth  int64   `json:"image_width"`
	ImageHeight int64   `json:"image_height"`
}

// ExportFrame is layout information for a single exported frame.
type ExportFrame struct {
	FrameNumber  int32   `json:"frame_number"`
	Bounds       []int64 `json:"bounds"`
	ImageWidth   int64   `json:"image_width"`
	ImageHeight  int64   `json:"image_height"`
	OffsetX      *int64  `json:"offset_x"`
	OffsetY      *int64  `json:"offset_y"`
	ScaledWidth  *int64  `json:"scaled_width"`
	ScaledHeight *int64  `json:"scaled_height"`
}

// TexturePackerExport is the JSON body returned by the texturepacker export.
type TexturePackerExport struct {
	Plan          ExportPlan      `json:"plan"`
	TexturePacker json.RawMessage `json:"texturepacker"`
	ImageBase64   string          `json:"image_base64"`
	ImageFilename string          `json:"image_filename"`
	JSONFilename  string          `json:"json_filename"`
}

// GodotExport is the JSON body returned by the godot export.
type GodotExport struct {
	Plan          ExportPlan `json:"plan"`
	Tres          string     `json:"tres"`
	ImageBase64   string     `json:"image_base64"`
	ImageFilename string     `json:"image_filename"`
	TresFilename  string     `json:"tres_filename"`
}
