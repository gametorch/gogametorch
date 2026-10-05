package gametorch

import (
	"encoding/json"
	"testing"
)

func TestReservationUsesBooleanCreated(t *testing.T) {
	const payload = `{
		"id": "b58cc74c-ea81-4a9b-b418-d785192011a4",
		"status": "queued",
		"created": true,
		"reserved_credits": "300.000000000000"
	}`
	var job JobCreated
	if err := json.Unmarshal([]byte(payload), &job); err != nil {
		t.Fatal(err)
	}
	if !job.Created {
		t.Error("created = false")
	}
	if job.ReservedCredits.String() != "300.000000000000" {
		t.Errorf("reserved_credits = %s", job.ReservedCredits.String())
	}
}

func TestSpriteGenerationToleratesIntAssetsAndArraySuggestions(t *testing.T) {
	const payload = `{
		"id": "b58cc74c-ea81-4a9b-b418-d785192011a4",
		"project_id": "afb90af1-80a4-4c75-9328-2de4edde4b16",
		"prompt": "fauna you'd find in a Western video game",
		"mode": "multiple",
		"image_model": "black-forest-labs/flux-3-image",
		"text_model": "openai/gpt-6-luna",
		"quality": null,
		"resolution": "1K",
		"base_asset_id": null,
		"status": "succeeded",
		"credits_consumed": "5.003952358800",
		"reserved_credits": "0.000000000000",
		"assets_delivered": 4,
		"archived_assets": 0,
		"error": null,
		"label_suggestions": ["enemy", "fauna"],
		"art_style_suggestion": null,
		"created_at": "2026-10-01T21:10:48.237007+00:00",
		"completed_at": "2026-10-01T21:11:22.052778+00:00",
		"archived_at": null,
		"user_id": "user_3JvOvirRbpZA5wZsIh3wgoRqYCQ",
		"source": "API key",
		"api_key_id": "93d6fe0b-b513-45cc-940c-2e8dc168f130",
		"key_name": "ci",
		"assets": [
			{
				"id": "857f82f2-f5a7-445f-a97d-a30097434324",
				"name": null,
				"width": 473,
				"height": 342,
				"archived_at": null,
				"has_original": true,
				"metadata": {},
				"labels": ["fauna"],
				"dismissed_suggestions": [],
				"generation_id": null,
				"project_id": null,
				"created_at": null,
				"prompt": null,
				"user_id": "user_abc",
				"source": "API key",
				"api_key_id": "93d6fe0b-b513-45cc-940c-2e8dc168f130",
				"key_name": null
			}
		]
	}`
	var generation Generation
	if err := json.Unmarshal([]byte(payload), &generation); err != nil {
		t.Fatal(err)
	}
	if generation.AssetsDelivered.Int64() != 4 {
		t.Errorf("assets_delivered = %d", generation.AssetsDelivered.Int64())
	}
	if len(generation.LabelSuggestions) != 2 {
		t.Errorf("label_suggestions = %v", generation.LabelSuggestions)
	}
	if len(generation.Assets) != 1 || generation.Assets[0].Width != 473 {
		t.Fatalf("assets = %+v", generation.Assets)
	}
	if !generation.Assets[0].HasOriginal {
		t.Error("has_original = false")
	}
	if generation.Provenance.Source != "API key" {
		t.Errorf("source = %q", generation.Provenance.Source)
	}
	if generation.Provenance.UserID != "user_3JvOvirRbpZA5wZsIh3wgoRqYCQ" {
		t.Errorf("user_id = %q", generation.Provenance.UserID)
	}
	if generation.Provenance.APIKeyID == nil {
		t.Error("api_key_id is nil")
	}
	if generation.Assets[0].Provenance.UserID != "user_abc" {
		t.Errorf("asset user_id = %q", generation.Assets[0].Provenance.UserID)
	}
}

func TestSpriteCatalogDecodes(t *testing.T) {
	const payload = `{
		"default_text_model": "openai/gpt-6-luna",
		"empirical_evidence": "v2/notes/sprite-provider-costs.md",
		"image_models": [
			{
				"id": "openai/gpt-image-2.5-flare",
				"name": "GPT Image 2.5 Flare",
				"blurb": "Best suited for base generation.",
				"available": true,
				"unavailable_reason": null,
				"qualities": ["auto", "low", "medium", "high"],
				"resolutions": [],
				"native_transparency": true,
				"editing": false,
				"transparency_status": "empirically_verified",
				"default_quality": "medium",
				"default_resolution": null,
				"default_canvas_size": "1024x1024",
				"capabilities_url": "https://openrouter.ai/api/v1/images/models/openai/gpt-image-2.5-flare/endpoints",
				"reservation_credits": "5.000000000000"
			}
		],
		"modes": ["single", "multiple"],
		"multiple_layout": {"rows": 2, "columns": 2, "images_per_request": 1},
		"reservation_credits": 300,
		"resolution_note": "Resolution describes the entire canvas.",
		"resolution_scope": "whole_canvas",
		"text_models": [{"id": "none", "name": "No prompt enhancement"}],
		"verified_at": "2026-09-27"
	}`
	var catalog SpriteModels
	if err := json.Unmarshal([]byte(payload), &catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog.ImageModels) != 1 {
		t.Fatalf("image_models = %d", len(catalog.ImageModels))
	}
	if catalog.ReservationCredits == nil || catalog.ReservationCredits.String() != "300" {
		t.Errorf("reservation_credits = %v", catalog.ReservationCredits)
	}
	if catalog.MultipleLayout.Columns != 2 {
		t.Errorf("columns = %d", catalog.MultipleLayout.Columns)
	}
}

func TestAnimationRunDecodes(t *testing.T) {
	const payload = `{
		"id": "29dc50d6-f8da-44c7-afa2-c2f5dcbd3cbb",
		"project_id": "afb90af1-80a4-4c75-9328-2de4edde4b16",
		"prompt": "show him draw his pistol and aim to the right",
		"animation_model": "ash",
		"duration": 4,
		"animation": {
			"id": "23d05f16-46f0-4672-8562-a9b90a9a46fb",
			"duration_seconds": 4,
			"resolution": "720p",
			"created_at": "2026-10-01T02:06:52.153372+00:00"
		},
		"status": "succeeded",
		"credits_consumed": "73.347960000000",
		"reserved_credits": "0.000000000000",
		"assets_delivered": 1,
		"base_asset_id": "6021352a-ab35-4a17-960b-71d30adf53a0",
		"error": null,
		"created_at": "2026-10-01T02:04:15.744412+00:00",
		"completed_at": "2026-10-01T02:06:52.157269+00:00",
		"archived_at": null,
		"frame_runs": [
			{
				"id": "3241879e-833f-4dee-b6a6-a19661102893",
				"status": "succeeded",
				"fps": 12,
				"frame_count": 49,
				"credits_consumed": "3.282263405160",
				"reserved_credits": "0.000000000000",
				"error": null,
				"created_at": "2026-10-01T02:06:52.155037+00:00"
			}
		],
		"frames": [
			{
				"id": "a2efe498-fa0b-4a65-a2a1-d24edd554954",
				"frame_number": 1,
				"generation_id": "3241879e-833f-4dee-b6a6-a19661102893",
				"created_at": "2026-10-01T02:06:55.498587+00:00"
			}
		]
	}`
	var run AnimationRun
	if err := json.Unmarshal([]byte(payload), &run); err != nil {
		t.Fatal(err)
	}
	if run.AnimationModel == nil || *run.AnimationModel != "ash" {
		t.Errorf("animation_model = %v", run.AnimationModel)
	}
	if run.Animation == nil || run.Animation.Resolution != "720p" {
		t.Errorf("animation = %+v", run.Animation)
	}
	if run.FrameRuns[0].FPS != 12 {
		t.Errorf("fps = %d", run.FrameRuns[0].FPS)
	}
	if run.Frames[0].FrameNumber != 1 {
		t.Errorf("frame_number = %d", run.Frames[0].FrameNumber)
	}
}

func TestUsageHistogramDecodesSources(t *testing.T) {
	const payload = `{
		"range": "24h",
		"width_seconds": 3600,
		"sources": [{"id": "api_key:abc", "label": "My Key", "kind": "api_key"}],
		"buckets": [
			{
				"start": "2026-10-04T01:00:00Z",
				"end": "2026-10-04T02:00:00Z",
				"total": "0",
				"values": {"api_key:abc": "0"}
			}
		]
	}`
	var histogram UsageHistogram
	if err := json.Unmarshal([]byte(payload), &histogram); err != nil {
		t.Fatal(err)
	}
	if histogram.Sources[0].ID != "api_key:abc" {
		t.Errorf("source id = %q", histogram.Sources[0].ID)
	}
	if histogram.Buckets[0].Total.String() != "0" {
		t.Errorf("total = %s", histogram.Buckets[0].Total.String())
	}
	if histogram.Buckets[0].Values["api_key:abc"].String() != "0" {
		t.Errorf("values = %v", histogram.Buckets[0].Values)
	}
}

func TestExportPlanDecodes(t *testing.T) {
	const payload = `{
		"start_frame": 1,
		"end_frame": 4,
		"reference": {"bounds": [2, 2, 433, 974], "image_width": 437, "image_height": 978},
		"first_frame": {"frame_number": 1, "image_width": 720, "image_height": 1280, "bounds": [115, 89, 490, 1102]},
		"frames": [
			{
				"frame_number": 1,
				"image_width": 720,
				"image_height": 1280,
				"bounds": [115, 89, 490, 1102],
				"offset_x": 0,
				"offset_y": 1,
				"scaled_width": 433,
				"scaled_height": 974
			}
		],
		"max_width": 490,
		"max_height": 1104,
		"scale": 0.8838475499092558,
		"scale_width": 0.8836734693877552,
		"scale_height": 0.8838475499092558,
		"canvas_width": 433,
		"canvas_height": 976,
		"frame_count": 4
	}`
	var plan ExportPlan
	if err := json.Unmarshal([]byte(payload), &plan); err != nil {
		t.Fatal(err)
	}
	if plan.FrameCount != 4 {
		t.Errorf("frame_count = %d", plan.FrameCount)
	}
	if plan.FirstFrame == nil || plan.FirstFrame.FrameNumber != 1 {
		t.Errorf("first_frame = %+v", plan.FirstFrame)
	}
	if plan.Frames[0].ScaledWidth == nil || *plan.Frames[0].ScaledWidth != 433 {
		t.Errorf("scaled_width = %v", plan.Frames[0].ScaledWidth)
	}
}

func TestLabelItemsDecodes(t *testing.T) {
	const payload = `{
		"label": {
			"id": "7fb262e8-6697-4d3f-b055-073874ce786e",
			"name": "enemy",
			"color": "#e57b7b",
			"thumbnail_asset_id": "ebe4fcb6-abaf-4a9e-84d6-b460d87e10a6",
			"created_at": "2026-09-28T21:07:15.245443Z"
		},
		"assets": [
			{
				"id": "7cf2e92f-5fe8-420e-8684-75017f01f766",
				"width": 955,
				"height": 1001,
				"archived_at": null,
				"name": null,
				"has_original": true,
				"metadata": {},
				"generation_id": "c4edd487-1b17-4d9a-b9cd-c6cfed1cfe9d",
				"prompt": "make its eyes solid black",
				"created_at": "2026-10-01T18:11:15.153977+00:00",
				"labels": ["enemy", "fauna"]
			}
		],
		"sounds": [],
		"saved_animations": []
	}`
	var items LabelItems
	if err := json.Unmarshal([]byte(payload), &items); err != nil {
		t.Fatal(err)
	}
	if items.Label.Name != "enemy" {
		t.Errorf("label name = %q", items.Label.Name)
	}
	if items.Assets[0].Width != 955 {
		t.Errorf("asset width = %d", items.Assets[0].Width)
	}
}

func TestArchiveResponseDecodes(t *testing.T) {
	const payload = `{"ok": true, "archived_at": "2026-09-29T19:49:33.368293+00:00"}`
	var response ArchiveResponse
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		t.Fatal(err)
	}
	if !response.OK || response.ArchivedAt == nil {
		t.Errorf("response = %+v", response)
	}
}

func TestLabelAssociationDecodes(t *testing.T) {
	const payload = `{"ok": true, "labels": ["character", "cowboy"]}`
	var association LabelAssociation
	if err := json.Unmarshal([]byte(payload), &association); err != nil {
		t.Fatal(err)
	}
	if len(association.Labels) != 2 || association.Labels[0] != "character" {
		t.Errorf("labels = %v", association.Labels)
	}
}

func TestApiKeyScopesDecode(t *testing.T) {
	const read = `{
		"id": "11111111-1111-1111-1111-111111111111",
		"name": "Read-only CI key",
		"key_prefix": "gt2_abc",
		"expires_at": null,
		"max_spend_limit": "0",
		"spend_reset_cadence": "monthly",
		"key_scope": "project_read",
		"project_id": "22222222-2222-2222-2222-222222222222",
		"spend": "0.000000000000",
		"lifetime_spend": "0.000000000000",
		"spend_reset_at": null,
		"created_at": "2026-10-05T00:00:00Z"
	}`
	var key ApiKey
	if err := json.Unmarshal([]byte(read), &key); err != nil {
		t.Fatal(err)
	}
	if key.KeyScope != ApiKeyScopeProjectRead {
		t.Errorf("scope = %q", key.KeyScope)
	}
	if key.ProjectID == nil || *key.ProjectID != "22222222-2222-2222-2222-222222222222" {
		t.Errorf("project_id = %v", key.ProjectID)
	}

	const admin = `{
		"id": "44444444-4444-4444-4444-444444444444",
		"key_prefix": "gt2_ghi",
		"spend_reset_cadence": "never",
		"key_scope": "admin",
		"project_id": null,
		"spend": "0",
		"lifetime_spend": "0",
		"created_at": "2026-10-05T00:00:00Z"
	}`
	if err := json.Unmarshal([]byte(admin), &key); err != nil {
		t.Fatal(err)
	}
	if key.KeyScope != ApiKeyScopeAdmin {
		t.Errorf("scope = %q", key.KeyScope)
	}
	if key.ProjectID != nil {
		t.Errorf("project_id = %v", key.ProjectID)
	}
}

func TestCreateApiKeyRequestSerialization(t *testing.T) {
	req := ProjectReadCreateApiKeyRequest("22222222-2222-2222-2222-222222222222").
		WithName("CI read key").
		WithMaxSpendLimit(MustDecimal("0")).
		WithSpendResetCadence(SpendResetMonthly)
	encoded, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["key_scope"] != "project_read" {
		t.Errorf("key_scope = %v", decoded["key_scope"])
	}
	if decoded["max_spend_limit"] != "0" {
		t.Errorf("max_spend_limit = %v", decoded["max_spend_limit"])
	}
	if decoded["spend_reset_cadence"] != "monthly" {
		t.Errorf("spend_reset_cadence = %v", decoded["spend_reset_cadence"])
	}
	if _, ok := decoded["expires_at"]; ok {
		t.Error("expires_at should be omitted when unset")
	}
}
