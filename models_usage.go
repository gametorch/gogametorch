package gametorch

import "time"

// Usage is the response from GET /usage.
type Usage struct {
	BalanceCredits  *Decimal       `json:"balance_credits"`
	ReservedCredits *Decimal       `json:"reserved_credits"`
	Summary         []UsageSummary `json:"summary"`
	Records         []UsageRecord  `json:"records"`
	NextCursor      *string        `json:"next_cursor"`
	Total           int64          `json:"total"`
}

// UsageSummary is a per-source spend summary entry.
type UsageSummary struct {
	Source          string  `json:"source"`
	APIKeyID        *string `json:"api_key_id"`
	UserID          *string `json:"user_id"`
	KeyName         *string `json:"key_name"`
	CreditsConsumed Decimal `json:"credits_consumed"`
}

// UsageRecord is a single operation-log entry.
type UsageRecord struct {
	ID                   string    `json:"id"`
	GenerationID         *string   `json:"generation_id"`
	CreatedAt            time.Time `json:"created_at"`
	Source               string    `json:"source"`
	Operation            *string   `json:"operation"`
	APIKeyID             *string   `json:"api_key_id"`
	UserID               *string   `json:"user_id"`
	Model                *string   `json:"model"`
	CreditsConsumed      Decimal   `json:"credits_consumed"`
	Settled              *bool     `json:"settled"`
	GenerationStatus     *string   `json:"generation_status"`
	Kind                 *string   `json:"kind"`
	ImageModel           *string   `json:"image_model"`
	AnimationModel       *string   `json:"animation_model"`
	ParentAnimationModel *string   `json:"parent_animation_model"`
}

// UsageHistogram is the response from GET /usage/histogram.
type UsageHistogram struct {
	Range        string                 `json:"range"`
	WidthSeconds int64                  `json:"width_seconds"`
	Sources      []UsageHistogramSource `json:"sources"`
	Buckets      []HistogramBucket      `json:"buckets"`
}

// UsageHistogramSource is a spend source in a usage histogram.
type UsageHistogramSource struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Kind  string `json:"kind"`
}

// HistogramBucket is a single histogram bucket.
type HistogramBucket struct {
	Start  time.Time  `json:"start"`
	End    time.Time  `json:"end"`
	Total  Decimal    `json:"total"`
	Values DecimalMap `json:"values"`
}
