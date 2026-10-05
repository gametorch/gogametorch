package gametorch

import (
	"context"
	"net/http"
)

type rangeBody struct {
	StartFrame int32 `json:"start_frame"`
	EndFrame   int32 `json:"end_frame"`
}

// ExportPlan returns the export plan (frame rectangles and metadata) without
// building a file.
//
// POST /animation-runs/{id}/export-plan
func (c *Client) ExportPlan(ctx context.Context, runID string, startFrame, endFrame int32) (*ExportPlan, error) {
	out, err := fetchJSON[ExportPlan](c, ctx, rateTier1, "POST /animation-runs/{id}/export-plan", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "animation-runs/" + runID + "/export-plan",
		body:   rangeBody{StartFrame: startFrame, EndFrame: endFrame},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Export builds and returns an export in the requested format. The result is a
// JSON document for ExportTexturePacker and ExportGodot, and raw bytes for
// every other format.
//
// POST /animation-runs/{id}/export/{format}
func (c *Client) Export(ctx context.Context, runID string, format ExportFormat, startFrame, endFrame int32) (*Export, error) {
	switch format {
	case ExportTexturePacker:
		out, err := c.ExportTexturePacker(ctx, runID, startFrame, endFrame)
		if err != nil {
			return nil, err
		}
		return &Export{TexturePacker: out}, nil
	case ExportGodot:
		out, err := c.ExportGodot(ctx, runID, startFrame, endFrame)
		if err != nil {
			return nil, err
		}
		return &Export{Godot: out}, nil
	default:
		out, err := c.exportBinary(ctx, runID, format, startFrame, endFrame)
		if err != nil {
			return nil, err
		}
		return &Export{Binary: out}, nil
	}
}

func (c *Client) exportBinary(ctx context.Context, runID string, format ExportFormat, startFrame, endFrame int32) (*Download, error) {
	return fetchDownload(c, ctx, rateTier1, "POST /animation-runs/{id}/export/{format}", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "animation-runs/" + runID + "/export/" + format.String(),
		body:   rangeBody{StartFrame: startFrame, EndFrame: endFrame},
	})
}

// ExportTexturePacker exports a TexturePacker JSON atlas (returned as
// structured JSON).
//
// POST /animation-runs/{id}/export/texturepacker
func (c *Client) ExportTexturePacker(ctx context.Context, runID string, startFrame, endFrame int32) (*TexturePackerExport, error) {
	out, err := fetchJSON[TexturePackerExport](c, ctx, rateTier1, "POST /animation-runs/{id}/export/texturepacker", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "animation-runs/" + runID + "/export/texturepacker",
		body:   rangeBody{StartFrame: startFrame, EndFrame: endFrame},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ExportGodot exports a Godot .tres resource and PNG (returned as structured
// JSON).
//
// POST /animation-runs/{id}/export/godot
func (c *Client) ExportGodot(ctx context.Context, runID string, startFrame, endFrame int32) (*GodotExport, error) {
	out, err := fetchJSON[GodotExport](c, ctx, rateTier1, "POST /animation-runs/{id}/export/godot", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "animation-runs/" + runID + "/export/godot",
		body:   rangeBody{StartFrame: startFrame, EndFrame: endFrame},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ExportTexturePackerZip exports a zipped TexturePacker JSON atlas.
func (c *Client) ExportTexturePackerZip(ctx context.Context, runID string, startFrame, endFrame int32) (*Download, error) {
	return c.exportBinary(ctx, runID, ExportTexturePackerZip, startFrame, endFrame)
}

// ExportAseprite exports an Aseprite file.
func (c *Client) ExportAseprite(ctx context.Context, runID string, startFrame, endFrame int32) (*Download, error) {
	return c.exportBinary(ctx, runID, ExportAseprite, startFrame, endFrame)
}

// ExportGodotZip exports a zipped Godot .tres and PNG.
func (c *Client) ExportGodotZip(ctx context.Context, runID string, startFrame, endFrame int32) (*Download, error) {
	return c.exportBinary(ctx, runID, ExportGodotZip, startFrame, endFrame)
}

// ExportGrid exports a single grid-strip PNG.
func (c *Client) ExportGrid(ctx context.Context, runID string, startFrame, endFrame int32) (*Download, error) {
	return c.exportBinary(ctx, runID, ExportGrid, startFrame, endFrame)
}

// ExportGameMaker exports a GameMaker strip PNG.
func (c *Client) ExportGameMaker(ctx context.Context, runID string, startFrame, endFrame int32) (*Download, error) {
	return c.exportBinary(ctx, runID, ExportGameMaker, startFrame, endFrame)
}

// ExportSequenceZip exports a zipped numbered PNG sequence.
func (c *Client) ExportSequenceZip(ctx context.Context, runID string, startFrame, endFrame int32) (*Download, error) {
	return c.exportBinary(ctx, runID, ExportSequenceZip, startFrame, endFrame)
}
