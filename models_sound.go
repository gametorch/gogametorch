package gametorch

import "time"

// SoundGeneration is a sound generation and its results.
type SoundGeneration struct {
	ID               string       `json:"id"`
	ProjectID        string       `json:"project_id"`
	Prompt           string       `json:"prompt"`
	SoundModel       string       `json:"sound_model"`
	ResponseFormat   *string      `json:"response_format"`
	Status           string       `json:"status"`
	CreditsConsumed  Decimal      `json:"credits_consumed"`
	ReservedCredits  Decimal      `json:"reserved_credits"`
	AssetsDelivered  BoolOrInt    `json:"assets_delivered"`
	ArchivedAssets   int64        `json:"archived_assets"`
	Error            *string      `json:"error"`
	LabelSuggestions StringList   `json:"label_suggestions"`
	CreatedAt        time.Time    `json:"created_at"`
	CompletedAt      *time.Time   `json:"completed_at"`
	ArchivedAt       *time.Time   `json:"archived_at"`
	Assets           []SoundAsset `json:"assets"`
	Provenance
}

// SoundGenerationsResponse is the response from
// GET /projects/{project_id}/sound-generations.
type SoundGenerationsResponse struct {
	SoundGenerations []SoundGeneration `json:"sound_generations"`
	NextCursor       *string           `json:"next_cursor"`
	Total            int64             `json:"total"`
}

// SoundAsset is a sound asset.
type SoundAsset struct {
	ID                   string            `json:"id"`
	Format               string            `json:"format"`
	Name                 *string           `json:"name"`
	Labels               []string          `json:"labels"`
	Metadata             map[string]string `json:"metadata"`
	DismissedSuggestions StringList        `json:"dismissed_suggestions"`
	CreatedAt            *time.Time        `json:"created_at"`
	ArchivedAt           *time.Time        `json:"archived_at"`
	Provenance
}
