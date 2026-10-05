package gametorch

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// liveClient returns a client for the live suites, skipping the test when
// GAMETORCH_API_KEY is not set. Point the suites at a local deployment with
// GAMETORCH_BASE_URL=http://localhost:8300/api.
func liveClient(t *testing.T) *Client {
	t.Helper()
	apiKey := os.Getenv(EnvAPIKey)
	if apiKey == "" {
		t.Skip("skipping live test: GAMETORCH_API_KEY is not set")
	}
	opts := []Option{WithAPIKey(apiKey)}
	if baseURL := os.Getenv(EnvBaseURL); baseURL != "" {
		opts = append(opts, WithBaseURL(baseURL))
	}
	client, err := NewClient(opts...)
	if err != nil {
		t.Fatalf("build client: %v", err)
	}
	return client
}

func liveWritesEnabled(t *testing.T) {
	t.Helper()
	if os.Getenv("GAMETORCH_LIVE_WRITES") != "1" {
		t.Skip("skipping live write test: set GAMETORCH_LIVE_WRITES=1")
	}
}

func liveSpendEnabled(t *testing.T) {
	t.Helper()
	if os.Getenv("GAMETORCH_LIVE_SPEND") != "1" {
		t.Skip("skipping spend test: set GAMETORCH_LIVE_SPEND=1")
	}
}

func keepProject() bool {
	return os.Getenv("GAMETORCH_KEEP_PROJECT") == "1"
}

func shortName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, newUUIDv4()[:8])
}

func projectName(kind string) string {
	return fmt.Sprintf("SDK %s %s", kind, newUUIDv4()[:8])
}

func firstProject(ctx context.Context, client *Client) (string, bool) {
	projects, err := client.ListProjects(ctx)
	if err != nil || len(projects.Projects) == 0 {
		return "", false
	}
	return projects.Projects[0].ID, true
}

// createLiveProject creates a fresh project, or reports false when the key may
// not create projects.
func createLiveProject(ctx context.Context, t *testing.T, client *Client, kind string) (*Project, bool) {
	t.Helper()
	project, err := client.CreateProject(ctx, projectName(kind))
	if err != nil {
		if e := asError(err); e.IsForbidden() {
			t.Log("skipping: this key cannot create projects")
			return nil, false
		}
		t.Fatalf("create project: %v", err)
	}
	t.Logf("created project %q (%s)", project.Name, project.Slug)
	return project, true
}

func finishLiveProject(ctx context.Context, t *testing.T, client *Client, slug string) {
	t.Helper()
	if keepProject() {
		t.Logf("keeping project %q (GAMETORCH_KEEP_PROJECT=1)", slug)
		return
	}
	if _, err := client.DeleteProject(ctx, slug); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	t.Logf("deleted project %q", slug)
}

func TestLiveHealthIsOK(t *testing.T) {
	client := liveClient(t)
	status, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("health: %v", err)
	}
	if status == "" {
		t.Fatal("health status is empty")
	}
}

func TestLiveCatalogsDecode(t *testing.T) {
	client := liveClient(t)
	ctx := context.Background()

	sprites, err := client.SpriteModels(ctx)
	if err != nil {
		t.Fatalf("sprite models: %v", err)
	}
	if len(sprites.ImageModels) == 0 {
		t.Fatal("no image models")
	}

	sounds, err := client.SoundModels(ctx)
	if err != nil {
		t.Fatalf("sound models: %v", err)
	}
	if len(sounds.Formats) == 0 {
		t.Fatal("no sound formats")
	}

	animations, err := client.AnimationModels(ctx)
	if err != nil {
		t.Fatalf("animation models: %v", err)
	}
	if len(animations.Data) == 0 {
		t.Fatal("no animation models")
	}
}

func TestLiveListProjectsDecodes(t *testing.T) {
	client := liveClient(t)
	projects, err := client.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("projects: %v", err)
	}
	for _, project := range projects.Projects {
		if project.Slug == "" {
			t.Errorf("project %s has an empty slug", project.ID)
		}
	}
}

func TestLiveUsageDecodes(t *testing.T) {
	client := liveClient(t)
	ctx := context.Background()

	usage, err := client.Usage(ctx, nil, nil)
	if err != nil {
		t.Fatalf("usage: %v", err)
	}
	if usage.Total < 0 {
		t.Errorf("total = %d", usage.Total)
	}

	histogram, err := client.UsageHistogram(ctx, Ptr("24h"), nil, nil)
	if err != nil {
		t.Fatalf("histogram: %v", err)
	}
	if histogram.Range != "24h" {
		t.Errorf("range = %q", histogram.Range)
	}
}

func TestLiveGenerationsAndAssetsDecode(t *testing.T) {
	client := liveClient(t)
	ctx := context.Background()
	projectID, ok := firstProject(ctx, client)
	if !ok {
		t.Skip("skipping: no projects")
	}

	params := NewListParams().WithIncludeArchived(true)
	generations, err := client.ListGenerations(ctx, projectID, params)
	if err != nil {
		t.Fatalf("generations: %v", err)
	}
	for i, generation := range generations.Generations {
		if i >= 3 {
			break
		}
		fetched, err := client.GetGeneration(ctx, generation.ID, true)
		if err != nil {
			t.Fatalf("get generation: %v", err)
		}
		if fetched.ID != generation.ID {
			t.Errorf("id = %s, want %s", fetched.ID, generation.ID)
		}
		if len(fetched.Assets) == 0 {
			continue
		}
		asset, err := client.GetAsset(ctx, fetched.Assets[0].ID)
		if err != nil {
			t.Fatalf("get asset: %v", err)
		}
		content, err := client.AssetContent(ctx, asset.ID)
		if err != nil {
			t.Fatalf("asset content: %v", err)
		}
		if content.IsEmpty() {
			t.Error("asset content is empty")
		}
		if asset.HasOriginal {
			original, err := client.AssetOriginal(ctx, asset.ID)
			if err != nil {
				t.Fatalf("asset original: %v", err)
			}
			if original.IsEmpty() {
				t.Error("asset original is empty")
			}
		}
	}

	assets, err := client.ListSpriteAssets(ctx, projectID, nil)
	if err != nil {
		t.Fatalf("sprite assets: %v", err)
	}
	for i, asset := range assets.Assets {
		if i >= 3 {
			break
		}
		if asset.Width < 0 {
			t.Errorf("asset %s width = %d", asset.ID, asset.Width)
		}
	}
}

func TestLiveSoundsDecode(t *testing.T) {
	client := liveClient(t)
	ctx := context.Background()
	projectID, ok := firstProject(ctx, client)
	if !ok {
		t.Skip("skipping: no projects")
	}

	params := NewListParams().WithIncludeArchived(true)
	sounds, err := client.ListSoundGenerations(ctx, projectID, params)
	if err != nil {
		t.Fatalf("sound generations: %v", err)
	}
	for i, generation := range sounds.SoundGenerations {
		if i >= 3 {
			break
		}
		fetched, err := client.GetSoundGeneration(ctx, generation.ID, true)
		if err != nil {
			t.Fatalf("get sound generation: %v", err)
		}
		if fetched.ID != generation.ID {
			t.Errorf("id = %s, want %s", fetched.ID, generation.ID)
		}
		if len(fetched.Assets) == 0 {
			continue
		}
		content, err := client.SoundAssetContent(ctx, fetched.Assets[0].ID)
		if err != nil {
			t.Fatalf("sound content: %v", err)
		}
		if content.IsEmpty() {
			t.Error("sound content is empty")
		}
	}
}

func TestLiveAnimationsAndExportsDecode(t *testing.T) {
	client := liveClient(t)
	ctx := context.Background()
	projectID, ok := firstProject(ctx, client)
	if !ok {
		t.Skip("skipping: no projects")
	}

	params := NewListParams().WithIncludeArchived(true)
	runs, err := client.ListAnimationRuns(ctx, projectID, params)
	if err != nil {
		t.Fatalf("animation runs: %v", err)
	}
	for i, run := range runs.Animations {
		if i >= 2 {
			break
		}
		fetched, err := client.GetAnimationRun(ctx, run.ID)
		if err != nil {
			t.Fatalf("get animation run: %v", err)
		}
		if fetched.ID != run.ID {
			t.Errorf("id = %s, want %s", fetched.ID, run.ID)
		}
		if fetched.Provenance.Source == "" {
			t.Error("animation run has no provenance source")
		}
		if len(fetched.Frames) > 0 {
			content, err := client.FrameContent(ctx, fetched.Frames[0].ID)
			if err != nil {
				t.Fatalf("frame content: %v", err)
			}
			if content.IsEmpty() {
				t.Error("frame content is empty")
			}
			plan, err := client.ExportPlan(ctx, run.ID, 1, 2)
			if err != nil {
				t.Fatalf("export plan: %v", err)
			}
			if plan.FrameCount < 0 {
				t.Errorf("frame_count = %d", plan.FrameCount)
			}
		}
	}

	base := ""
	for _, run := range runs.Animations {
		if run.BaseAssetID != nil {
			base = *run.BaseAssetID
			break
		}
	}
	if base != "" {
		filtered, err := client.ListAnimationRuns(ctx, projectID, NewListParams().WithIncludeArchived(true).WithBaseAssetID(base))
		if err != nil {
			t.Fatalf("filter animation runs: %v", err)
		}
		for _, run := range filtered.Animations {
			if run.BaseAssetID == nil || *run.BaseAssetID != base {
				t.Errorf("run %s base_asset_id = %v, want %s", run.ID, run.BaseAssetID, base)
			}
		}
	}
}

func TestLiveSavedAnimationsDecode(t *testing.T) {
	client := liveClient(t)
	ctx := context.Background()
	projectID, ok := firstProject(ctx, client)
	if !ok {
		t.Skip("skipping: no projects")
	}

	params := NewListParams().WithIncludeArchived(true)
	saved, err := client.ListSavedAnimations(ctx, projectID, params)
	if err != nil {
		t.Fatalf("saved animations: %v", err)
	}
	for i, animation := range saved.SavedAnimations {
		if i >= 3 {
			break
		}
		fetched, err := client.GetSavedAnimation(ctx, animation.ID)
		if err != nil {
			t.Fatalf("get saved animation: %v", err)
		}
		if fetched.ID != animation.ID {
			t.Errorf("id = %s, want %s", fetched.ID, animation.ID)
		}
	}
}

func TestLiveLabelsAndArtStylesDecode(t *testing.T) {
	client := liveClient(t)
	ctx := context.Background()
	projectID, ok := firstProject(ctx, client)
	if !ok {
		t.Skip("skipping: no projects")
	}

	labels, err := client.ListLabels(ctx, projectID)
	if err != nil {
		t.Fatalf("labels: %v", err)
	}
	if len(labels.Labels) > 0 {
		items, err := client.LabelItems(ctx, labels.Labels[0].ID)
		if err != nil {
			t.Fatalf("label items: %v", err)
		}
		if items.Label.ID != labels.Labels[0].ID {
			t.Errorf("label id = %s, want %s", items.Label.ID, labels.Labels[0].ID)
		}
	}

	styles, err := client.ListArtStyles(ctx, projectID)
	if err != nil {
		t.Fatalf("art styles: %v", err)
	}
	for _, style := range styles.ArtStyles {
		if style.Name == "" {
			t.Errorf("art style %s has an empty name", style.ID)
		}
	}
}

func TestLiveKeysDecodeOrForbidden(t *testing.T) {
	client := liveClient(t)
	keys, err := client.ListKeys(context.Background())
	if err != nil {
		if e := asError(err); e.IsForbidden() {
			return
		}
		t.Fatalf("keys: %v", err)
	}
	for _, key := range keys.Keys {
		if key.KeyPrefix == "" {
			t.Errorf("key %s has an empty prefix", key.ID)
		}
	}
}

func TestLiveAnimationEstimateDecodes(t *testing.T) {
	client := liveClient(t)
	ctx := context.Background()
	projectID, ok := firstProject(ctx, client)
	if !ok {
		t.Skip("skipping: no projects")
	}

	estimate, err := client.EstimateAnimation(projectID).
		WithAnimationModel("ash").
		WithDuration(4).
		Send(ctx)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	if estimate.AnimationModel != "ash" {
		t.Errorf("animation_model = %q", estimate.AnimationModel)
	}
	if estimate.Duration != 4 {
		t.Errorf("duration = %d", estimate.Duration)
	}
	if estimate.Resolution == "" {
		t.Error("resolution is empty")
	}
}

func TestLiveUnauthenticatedRequestsAreRejected(t *testing.T) {
	if os.Getenv(EnvAPIKey) == "" {
		t.Skip("skipping live test: GAMETORCH_API_KEY is not set")
	}
	opts := []Option{}
	if baseURL := os.Getenv(EnvBaseURL); baseURL != "" {
		opts = append(opts, WithBaseURL(baseURL))
	}
	anonymous, err := NewClient(opts...)
	if err != nil {
		t.Fatal(err)
	}
	_, err = anonymous.ListProjects(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	e := asError(err)
	if !e.IsUnauthorized() {
		t.Fatalf("expected 401, got kind %q status %d", e.Kind(), e.StatusCode())
	}
}
