package gametorch

import "time"

// Label is a project label.
type Label struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Color            *string   `json:"color"`
	ThumbnailAssetID *string   `json:"thumbnail_asset_id"`
	CreatedAt        time.Time `json:"created_at"`
}

// LabelsResponse is the response from GET /projects/{project_id}/labels.
type LabelsResponse struct {
	Labels []Label `json:"labels"`
}

// LabelAssociation is the label-name set returned by a label mutation.
type LabelAssociation struct {
	OK     bool     `json:"ok"`
	Labels []string `json:"labels"`
}

// LabelItems is the response from GET /labels/{label_id}/items.
type LabelItems struct {
	Label           Label                 `json:"label"`
	Assets          []LabelAsset          `json:"assets"`
	Sounds          []LabelSound          `json:"sounds"`
	SavedAnimations []LabelSavedAnimation `json:"saved_animations"`
}

// LabelAsset is a sprite asset as returned by the label-items endpoint.
type LabelAsset struct {
	ID           string            `json:"id"`
	Width        int64             `json:"width"`
	Height       int64             `json:"height"`
	Name         *string           `json:"name"`
	Labels       []string          `json:"labels"`
	Metadata     map[string]string `json:"metadata"`
	HasOriginal  bool              `json:"has_original"`
	GenerationID *string           `json:"generation_id"`
	Prompt       *string           `json:"prompt"`
	CreatedAt    *time.Time        `json:"created_at"`
	ArchivedAt   *time.Time        `json:"archived_at"`
	Provenance
}

// LabelSound is a sound asset as returned by the label-items endpoint.
type LabelSound struct {
	ID           string            `json:"id"`
	Format       string            `json:"format"`
	Name         *string           `json:"name"`
	Labels       []string          `json:"labels"`
	Metadata     map[string]string `json:"metadata"`
	GenerationID *string           `json:"generation_id"`
	Prompt       *string           `json:"prompt"`
	CreatedAt    *time.Time        `json:"created_at"`
	ArchivedAt   *time.Time        `json:"archived_at"`
	Provenance
}

// LabelSavedAnimation is a saved animation as returned by the label-items
// endpoint.
type LabelSavedAnimation struct {
	ID             string            `json:"id"`
	ProjectID      string            `json:"project_id"`
	GenerationID   string            `json:"generation_id"`
	StartFrame     int64             `json:"start_frame"`
	EndFrame       int64             `json:"end_frame"`
	Name           *string           `json:"name"`
	AnimationModel *string           `json:"animation_model"`
	Prompt         *string           `json:"prompt"`
	BaseAssetID    *string           `json:"base_asset_id"`
	FrameCount     int64             `json:"frame_count"`
	Labels         []string          `json:"labels"`
	Metadata       map[string]string `json:"metadata"`
	CreatedAt      *time.Time        `json:"created_at"`
	ArchivedAt     *time.Time        `json:"archived_at"`
	Provenance
}
