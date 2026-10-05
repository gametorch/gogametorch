package gametorch

import (
	"context"
	"testing"
	"time"
)

func TestLiveProjectLifecycle(t *testing.T) {
	liveWritesEnabled(t)
	client := liveClient(t)
	ctx := context.Background()

	project, ok := createLiveProject(ctx, t, client, "Project Test")
	if !ok {
		return
	}

	projects, err := client.ListProjects(ctx)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	found := false
	for _, p := range projects.Projects {
		if p.ID == project.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created project %s not in list", project.ID)
	}

	renamed, err := client.RenameProject(ctx, project.Slug, projectName("Project Test Renamed"))
	if err != nil {
		t.Fatalf("rename project: %v", err)
	}
	if renamed.ID != project.ID {
		t.Errorf("renamed id = %s, want %s", renamed.ID, project.ID)
	}

	finishLiveProject(ctx, t, client, renamed.Slug)
}

func TestLiveLabelLifecycle(t *testing.T) {
	liveWritesEnabled(t)
	client := liveClient(t)
	ctx := context.Background()

	project, ok := createLiveProject(ctx, t, client, "Label Test")
	if !ok {
		return
	}

	name := shortName("sdk-label")
	label, err := client.CreateLabel(ctx, project.ID, name, Ptr("#123456"))
	if err != nil {
		t.Fatalf("create label: %v", err)
	}
	if label.Name != name {
		t.Errorf("label name = %q, want %q", label.Name, name)
	}

	newName := shortName("sdk-renamed")
	updated, err := client.UpdateLabel(ctx, label.ID, Ptr(newName), Ptr("#654321"))
	if err != nil {
		t.Fatalf("update label: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("updated name = %q, want %q", updated.Name, newName)
	}

	labels, err := client.ListLabels(ctx, project.ID)
	if err != nil {
		t.Fatalf("list labels: %v", err)
	}
	found := false
	for _, l := range labels.Labels {
		if l.ID == label.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("label %s not in list", label.ID)
	}

	if _, err := client.DeleteLabel(ctx, label.ID); err != nil {
		t.Fatalf("delete label: %v", err)
	}

	kept, err := client.CreateLabel(ctx, project.ID, shortName("sdk-kept"), Ptr("#8bc34a"))
	if err != nil {
		t.Fatalf("create kept label: %v", err)
	}
	t.Logf("kept label %q", kept.Name)

	finishLiveProject(ctx, t, client, project.Slug)
}

func TestLiveArtStyleLifecycle(t *testing.T) {
	liveWritesEnabled(t)
	client := liveClient(t)
	ctx := context.Background()

	project, ok := createLiveProject(ctx, t, client, "Art Style Test")
	if !ok {
		return
	}

	name := shortName("sdk-style")
	style, err := client.CreateArtStyle(ctx, project.ID, name)
	if err != nil {
		t.Fatalf("create art style: %v", err)
	}
	if style.Name != name {
		t.Errorf("style name = %q, want %q", style.Name, name)
	}

	styles, err := client.ListArtStyles(ctx, project.ID)
	if err != nil {
		t.Fatalf("list styles: %v", err)
	}
	found := false
	for _, s := range styles.ArtStyles {
		if s.ID == style.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("art style %s not in list", style.ID)
	}

	if _, err := client.DeleteArtStyle(ctx, style.ID); err != nil {
		t.Fatalf("delete art style: %v", err)
	}

	kept, err := client.CreateArtStyle(ctx, project.ID, shortName("sdk-kept-style"))
	if err != nil {
		t.Fatalf("create kept art style: %v", err)
	}
	t.Logf("kept art style %q", kept.Name)

	finishLiveProject(ctx, t, client, project.Slug)
}

func TestLiveGenerateArtStyleDecodes(t *testing.T) {
	liveWritesEnabled(t)
	client := liveClient(t)
	ctx := context.Background()

	project, ok := createLiveProject(ctx, t, client, "Art Style Suggest Test")
	if !ok {
		return
	}

	suggestion, err := client.GenerateArtStyle(ctx, project.ID)
	if err != nil {
		t.Fatalf("generate art style: %v", err)
	}
	if suggestion.Name == "" {
		t.Error("suggestion name is empty")
	}

	finishLiveProject(ctx, t, client, project.Slug)
}

func TestLiveEnsureUserOK(t *testing.T) {
	liveWritesEnabled(t)
	client := liveClient(t)

	response, err := client.EnsureUser(context.Background())
	if err != nil {
		t.Fatalf("ensure user: %v", err)
	}
	if !response.OK {
		t.Error("ensure user returned ok=false")
	}
}

func TestLiveKeyLifecycle(t *testing.T) {
	liveWritesEnabled(t)
	client := liveClient(t)
	ctx := context.Background()

	created, err := client.CreateKey(ctx, NewCreateApiKeyRequest().WithName(shortName("sdk-key")))
	if err != nil {
		if e := asError(err); e.IsForbidden() {
			t.Log("skipping: this key cannot manage API keys (admin required)")
			return
		}
		t.Fatalf("create key: %v", err)
	}
	if created.KeyFull == "" {
		t.Error("key_full is empty")
	}
	if created.KeyPrefix == "" {
		t.Error("key_prefix is empty")
	}

	updated, err := client.UpdateKey(ctx, created.ID, NewUpdateApiKeyRequest().WithName(shortName("sdk-key-renamed")))
	if err != nil {
		t.Fatalf("update key: %v", err)
	}
	if updated.Name == nil {
		t.Error("updated name is nil")
	}

	if _, err := client.DeleteKey(ctx, created.ID); err != nil {
		t.Fatalf("delete key: %v", err)
	}
}

func TestLiveScopedKeyLifecycle(t *testing.T) {
	liveWritesEnabled(t)
	client := liveClient(t)
	ctx := context.Background()

	project, ok := createLiveProject(ctx, t, client, "Scoped Key Test")
	if !ok {
		return
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	read, err := client.CreateKey(ctx, ProjectReadCreateApiKeyRequest(project.ID).
		WithName(shortName("sdk-read")).
		WithMaxSpendLimit(MustDecimal("0")).
		WithSpendResetCadence(SpendResetMonthly).
		WithExpiresAt(expiresAt))
	if err != nil {
		if e := asError(err); e.IsForbidden() {
			t.Log("skipping: this key cannot manage API keys (admin required)")
			finishLiveProject(ctx, t, client, project.Slug)
			return
		}
		t.Fatalf("create read key: %v", err)
	}
	if read.KeyScope != ApiKeyScopeProjectRead {
		t.Errorf("read scope = %q", read.KeyScope)
	}
	if read.ProjectID == nil || *read.ProjectID != project.ID {
		t.Errorf("read project = %v", read.ProjectID)
	}
	if read.Name == nil || read.ExpiresAt == nil {
		t.Error("read key name/expiry missing")
	}

	write, err := client.CreateKey(ctx, ProjectWriteCreateApiKeyRequest(project.ID).
		WithName(shortName("sdk-write")).
		WithMaxSpendLimit(MustDecimal("500")).
		WithSpendResetCadence(SpendResetWeekly).
		WithExpiresAt(expiresAt))
	if err != nil {
		t.Fatalf("create write key: %v", err)
	}
	if write.KeyScope != ApiKeyScopeProjectWrite {
		t.Errorf("write scope = %q", write.KeyScope)
	}
	if write.MaxSpendLimit == nil || write.MaxSpendLimit.String() != "500" {
		t.Errorf("write max_spend_limit = %v", write.MaxSpendLimit)
	}

	updated, err := client.UpdateKey(ctx, write.ID, NewUpdateApiKeyRequest().
		WithName(shortName("sdk-write-renamed")).
		WithMaxSpendLimit(MustDecimal("750")).
		WithSpendResetCadence(SpendResetMonthly).
		WithExpiresAt(expiresAt.Add(30*24*time.Hour)))
	if err != nil {
		t.Fatalf("update write key: %v", err)
	}
	if updated.MaxSpendLimit == nil || updated.MaxSpendLimit.String() != "750" {
		t.Errorf("updated max_spend_limit = %v", updated.MaxSpendLimit)
	}
	if updated.SpendResetCadence != "monthly" {
		t.Errorf("updated cadence = %q", updated.SpendResetCadence)
	}

	if _, err := client.DeleteKey(ctx, read.ID); err != nil {
		t.Fatalf("delete read key: %v", err)
	}
	if _, err := client.DeleteKey(ctx, write.ID); err != nil {
		t.Fatalf("delete write key: %v", err)
	}

	finishLiveProject(ctx, t, client, project.Slug)
}
