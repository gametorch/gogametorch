package gametorch

import "time"

// ArtStyle is a reusable project art style.
type ArtStyle struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ArtStylesResponse is the response from GET /projects/{project_id}/art-styles.
type ArtStylesResponse struct {
	ArtStyles []ArtStyle `json:"art_styles"`
}

// ArtStyleSuggestion is a freshly generated art-style suggestion.
type ArtStyleSuggestion struct {
	Name string `json:"name"`
}
