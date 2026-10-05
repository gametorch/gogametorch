package gametorch

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

func int64String(n int64) string { return strconv.FormatInt(n, 10) }

// OkResponse is a generic acknowledgement returned by delete and a few write
// endpoints.
type OkResponse struct {
	OK bool `json:"ok"`
}

// ArchiveResponse is returned by archive/unarchive endpoints.
type ArchiveResponse struct {
	OK         bool       `json:"ok"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
}

// AssetNameResponse is returned by asset rename endpoints.
type AssetNameResponse struct {
	OK   bool    `json:"ok"`
	Name *string `json:"name"`
}

// AssetMetadataResponse is returned by metadata endpoints.
type AssetMetadataResponse struct {
	OK       bool              `json:"ok"`
	Metadata map[string]string `json:"metadata"`
}

// JobCreated is the job-creation envelope returned by generation and
// frame-generation endpoints.
type JobCreated struct {
	ID              string  `json:"id"`
	Status          string  `json:"status"`
	Created         bool    `json:"created"`
	ReservedCredits Decimal `json:"reserved_credits"`
}

// Download is raw bytes returned by a content endpoint, along with the
// response's content type.
type Download struct {
	// ContentType is the Content-Type reported by the server, if any.
	ContentType string
	// Data is the response body.
	Data []byte
}

// Len returns the length of the body in bytes.
func (d *Download) Len() int { return len(d.Data) }

// IsEmpty reports whether the body is empty.
func (d *Download) IsEmpty() bool { return len(d.Data) == 0 }

// Page is one page of a paginated list response.
type Page[T any] struct {
	Items      []T
	NextCursor *string
	Total      int64
}

type pageFetcher[T any] func(ctx context.Context, cursor *string) (Page[T], error)

// Paginator is a cursor over a paginated GameTorch list endpoint. It fetches
// pages on demand as items are consumed.
type Paginator[T any] struct {
	fetch    pageFetcher[T]
	cursor   *string
	finished bool
	buffer   []T
	total    *int64
}

func newPaginator[T any](fetch pageFetcher[T]) *Paginator[T] {
	return &Paginator[T]{fetch: fetch}
}

// NextPage fetches the next page. It returns ok=false when the cursor is
// exhausted.
func (p *Paginator[T]) NextPage(ctx context.Context) (items []T, ok bool, err error) {
	if p.finished {
		return nil, false, nil
	}
	page, err := p.fetch(ctx, p.cursor)
	if err != nil {
		p.finished = true
		return nil, false, err
	}
	total := page.Total
	p.total = &total
	p.cursor = page.NextCursor
	if p.cursor == nil {
		p.finished = true
	}
	if len(page.Items) == 0 && p.finished {
		return nil, false, nil
	}
	return page.Items, true, nil
}

// Next fetches the next item, transparently loading pages as needed. It
// returns ok=false when the cursor is exhausted.
func (p *Paginator[T]) Next(ctx context.Context) (item T, ok bool, err error) {
	for {
		if len(p.buffer) > 0 {
			item = p.buffer[0]
			p.buffer = p.buffer[1:]
			return item, true, nil
		}
		items, more, err := p.NextPage(ctx)
		if err != nil {
			return item, false, err
		}
		if !more {
			return item, false, nil
		}
		p.buffer = items
	}
}

// Total returns the total number of items reported by the API, once the first
// page has been fetched.
func (p *Paginator[T]) Total() (int64, bool) {
	if p.total == nil {
		return 0, false
	}
	return *p.total, true
}

// ListParams are the options shared by the paginated list endpoints. The zero
// value lists the first page without archived items.
type ListParams struct {
	// Before is the cursor returned by a previous page.
	Before *string
	// IncludeArchived includes archived items (defaults to false).
	IncludeArchived bool
	// BaseAssetID filters animation runs by their base sprite asset. It is only
	// used by Client.ListAnimationRuns and ignored by other list endpoints.
	BaseAssetID *string
}

// NewListParams returns empty list parameters.
func NewListParams() *ListParams { return &ListParams{} }

// Ptr returns a pointer to v. It is a convenience for the optional fields on
// requests and list parameters.
func Ptr[T any](v T) *T { return &v }

// WithBefore sets the pagination cursor.
func (p *ListParams) WithBefore(cursor string) *ListParams {
	p.Before = &cursor
	return p
}

// WithIncludeArchived includes or excludes archived items.
func (p *ListParams) WithIncludeArchived(include bool) *ListParams {
	p.IncludeArchived = include
	return p
}

// WithBaseAssetID filters animation runs by their base sprite asset.
func (p *ListParams) WithBaseAssetID(baseAssetID string) *ListParams {
	p.BaseAssetID = &baseAssetID
	return p
}

// SpriteMode is the sprite generation mode.
type SpriteMode string

// Sprite generation modes.
const (
	// SpriteModeSingle generates a single sprite.
	SpriteModeSingle SpriteMode = "single"
	// SpriteModeMultiple generates four sprites from one 2x2 sheet.
	SpriteModeMultiple SpriteMode = "multiple"
)

// ParseSpriteMode parses a sprite mode.
func ParseSpriteMode(value string) (SpriteMode, error) {
	switch SpriteMode(value) {
	case SpriteModeSingle, SpriteModeMultiple:
		return SpriteMode(value), nil
	default:
		return "", fmt.Errorf("unknown sprite mode: %s", value)
	}
}

// ExportFormat is a supported animation export format.
type ExportFormat string

// Export formats.
const (
	ExportTexturePacker    ExportFormat = "texturepacker"
	ExportTexturePackerZip ExportFormat = "texturepacker.zip"
	ExportAseprite         ExportFormat = "aseprite"
	ExportGodot            ExportFormat = "godot"
	ExportGodotZip         ExportFormat = "godot.zip"
	ExportGrid             ExportFormat = "grid"
	ExportGameMaker        ExportFormat = "gamemaker"
	ExportSequenceZip      ExportFormat = "sequence.zip"
)

// String returns the path segment used by the export endpoint.
func (f ExportFormat) String() string { return string(f) }

// ParseExportFormat parses an export format.
func ParseExportFormat(value string) (ExportFormat, error) {
	switch ExportFormat(value) {
	case ExportTexturePacker, ExportTexturePackerZip, ExportAseprite, ExportGodot,
		ExportGodotZip, ExportGrid, ExportGameMaker, ExportSequenceZip:
		return ExportFormat(value), nil
	default:
		return "", fmt.Errorf("unknown export format: %s", value)
	}
}

// ApiKeyScope is what an API key is allowed to do. The scope is fixed at
// creation.
type ApiKeyScope string

// API key scopes.
const (
	// ApiKeyScopeAdmin grants full access to the owning account/org.
	ApiKeyScopeAdmin ApiKeyScope = "admin"
	// ApiKeyScopeProjectWrite is bound to one project with write access.
	ApiKeyScopeProjectWrite ApiKeyScope = "project_write"
	// ApiKeyScopeProjectRead is bound to one project with read-only access.
	ApiKeyScopeProjectRead ApiKeyScope = "project_read"
)

// String returns the wire representation of the scope.
func (s ApiKeyScope) String() string { return string(s) }

// ParseApiKeyScope parses an API key scope.
func ParseApiKeyScope(value string) (ApiKeyScope, error) {
	switch ApiKeyScope(value) {
	case ApiKeyScopeAdmin, ApiKeyScopeProjectWrite, ApiKeyScopeProjectRead:
		return ApiKeyScope(value), nil
	default:
		return "", fmt.Errorf("unknown API key scope: %s", value)
	}
}

// SpendResetCadence is how often an API key's spend counter resets.
type SpendResetCadence string

// Spend reset cadences.
const (
	SpendResetDaily   SpendResetCadence = "daily"
	SpendResetWeekly  SpendResetCadence = "weekly"
	SpendResetMonthly SpendResetCadence = "monthly"
	SpendResetNever   SpendResetCadence = "never"
)

// String returns the wire representation of the cadence.
func (c SpendResetCadence) String() string { return string(c) }

// ParseSpendResetCadence parses a spend reset cadence.
func ParseSpendResetCadence(value string) (SpendResetCadence, error) {
	switch SpendResetCadence(value) {
	case SpendResetDaily, SpendResetWeekly, SpendResetMonthly, SpendResetNever:
		return SpendResetCadence(value), nil
	default:
		return "", fmt.Errorf("unknown spend reset cadence: %s", value)
	}
}

// Export is the result of Client.Export, which varies by format.
type Export struct {
	// TexturePacker is set for the texturepacker format.
	TexturePacker *TexturePackerExport
	// Godot is set for the godot format.
	Godot *GodotExport
	// Binary is set for every other format.
	Binary *Download
}

// IsBinary reports whether the export is a binary payload.
func (e *Export) IsBinary() bool { return e != nil && e.Binary != nil }

// AsBinary returns the binary payload, if this is a binary export.
func (e *Export) AsBinary() (*Download, bool) {
	if e != nil && e.Binary != nil {
		return e.Binary, true
	}
	return nil, false
}
