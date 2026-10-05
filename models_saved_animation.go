package gametorch

import "time"

// SavedAnimation is a named frame range taken from an animation run, reusable
// as a preset.
type SavedAnimation struct {
	ID             string            `json:"id"`
	ProjectID      string            `json:"project_id"`
	GenerationID   string            `json:"generation_id"`
	Name           *string           `json:"name"`
	AnimationModel *string           `json:"animation_model"`
	Prompt         *string           `json:"prompt"`
	BaseAssetID    *string           `json:"base_asset_id"`
	StartFrame     int64             `json:"start_frame"`
	EndFrame       int64             `json:"end_frame"`
	FrameCount     int64             `json:"frame_count"`
	Labels         []string          `json:"labels"`
	Metadata       map[string]string `json:"metadata"`
	CreatedAt      time.Time         `json:"created_at"`
	ArchivedAt     *time.Time        `json:"archived_at"`
}

// SavedAnimationsResponse is the response from
// GET /projects/{project_id}/saved-animations.
type SavedAnimationsResponse struct {
	SavedAnimations []SavedAnimation `json:"saved_animations"`
	NextCursor      *string          `json:"next_cursor"`
	Total           int64            `json:"total"`
}
