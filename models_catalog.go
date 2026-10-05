package gametorch

// SpriteModels is the response from GET /sprite-models.
type SpriteModels struct {
	DefaultTextModel   string         `json:"default_text_model"`
	EmpiricalEvidence  *string        `json:"empirical_evidence"`
	ImageModels        []ImageModel   `json:"image_models"`
	Modes              []string       `json:"modes"`
	MultipleLayout     MultipleLayout `json:"multiple_layout"`
	ReservationCredits *Decimal       `json:"reservation_credits"`
	ResolutionNote     *string        `json:"resolution_note"`
	ResolutionScope    *string        `json:"resolution_scope"`
	TextModels         []TextModel    `json:"text_models"`
	VerifiedAt         *string        `json:"verified_at"`
}

// MultipleLayout is the layout metadata for sprite multiple mode.
type MultipleLayout struct {
	Columns          int64 `json:"columns"`
	ImagesPerRequest int64 `json:"images_per_request"`
	Rows             int64 `json:"rows"`
}

// ImageModel is an image model available for sprite generation.
type ImageModel struct {
	Available          bool     `json:"available"`
	Blurb              string   `json:"blurb"`
	CapabilitiesURL    *string  `json:"capabilities_url"`
	DefaultCanvasSize  *string  `json:"default_canvas_size"`
	DefaultQuality     *string  `json:"default_quality"`
	DefaultResolution  *string  `json:"default_resolution"`
	Editing            bool     `json:"editing"`
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	NativeTransparency bool     `json:"native_transparency"`
	Qualities          []string `json:"qualities"`
	ReservationCredits *Decimal `json:"reservation_credits"`
	Resolutions        []string `json:"resolutions"`
	TransparencyStatus *string  `json:"transparency_status"`
	UnavailableReason  *string  `json:"unavailable_reason"`
}

// TextModel is a prompt-enhancement text model.
type TextModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SoundModels is the response from GET /sound-models.
type SoundModels struct {
	DefaultFormat string         `json:"default_format"`
	Formats       []string       `json:"formats"`
	Model         SoundModelInfo `json:"model"`
}

// SoundModelInfo is a sound generation model.
type SoundModelInfo struct {
	Blurb string `json:"blurb"`
	ID    string `json:"id"`
	Name  string `json:"name"`
}

// AnimationModels is the response from GET /animation-models.
type AnimationModels struct {
	Data []AnimationModelInfo `json:"data"`
}

// AnimationModelInfo is an animation model addressed by its public name.
type AnimationModelInfo struct {
	Description          string   `json:"description"`
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	SupportedDurations   []int64  `json:"supported_durations"`
	SupportedResolutions []string `json:"supported_resolutions"`
}
