package gametorch

import "time"

// Generation is a sprite generation and its results.
type Generation struct {
	ID                 string     `json:"id"`
	ProjectID          string     `json:"project_id"`
	Prompt             string     `json:"prompt"`
	Mode               string     `json:"mode"`
	ImageModel         string     `json:"image_model"`
	TextModel          *string    `json:"text_model"`
	Quality            *string    `json:"quality"`
	Resolution         *string    `json:"resolution"`
	BaseAssetID        *string    `json:"base_asset_id"`
	Status             string     `json:"status"`
	CreditsConsumed    Decimal    `json:"credits_consumed"`
	ReservedCredits    Decimal    `json:"reserved_credits"`
	AssetsDelivered    BoolOrInt  `json:"assets_delivered"`
	ArchivedAssets     int64      `json:"archived_assets"`
	Warming            bool       `json:"warming"`
	Error              *string    `json:"error"`
	LabelSuggestions   StringList `json:"label_suggestions"`
	ArtStyleSuggestion *string    `json:"art_style_suggestion"`
	CreatedAt          time.Time  `json:"created_at"`
	CompletedAt        *time.Time `json:"completed_at"`
	ArchivedAt         *time.Time `json:"archived_at"`
	Assets             []Asset    `json:"assets"`
	Provenance
}

// GenerationsResponse is the response from GET /projects/{project_id}/generations.
type GenerationsResponse struct {
	Generations []Generation `json:"generations"`
	NextCursor  *string      `json:"next_cursor"`
	Total       int64        `json:"total"`
}

// Asset is a sprite asset. Some fields are absent depending on the endpoint;
// nested assets in a generation omit GenerationID and CreatedAt.
type Asset struct {
	ID                   string            `json:"id"`
	Name                 *string           `json:"name"`
	Width                int64             `json:"width"`
	Height               int64             `json:"height"`
	Labels               []string          `json:"labels"`
	Metadata             map[string]string `json:"metadata"`
	HasOriginal          bool              `json:"has_original"`
	DismissedSuggestions StringList        `json:"dismissed_suggestions"`
	GenerationID         *string           `json:"generation_id"`
	ProjectID            *string           `json:"project_id"`
	Prompt               *string           `json:"prompt"`
	CreatedAt            *time.Time        `json:"created_at"`
	ArchivedAt           *time.Time        `json:"archived_at"`
	Provenance
}

// SpriteAssetsResponse is the response from GET /projects/{project_id}/sprite-assets.
type SpriteAssetsResponse struct {
	Assets []Asset `json:"assets"`
}
