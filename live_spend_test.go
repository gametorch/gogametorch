package gametorch

import (
	"context"
	"testing"
	"time"
)

const (
	spriteFilepath    = "/foo/bar/sprites/hero.png"
	soundFilepath     = "/foo/bar/audio/sword-unsheath.mp3"
	animationFilepath = "/foo/bar/animations/sword-raise.aseprite"
)

var allFormats = []ExportFormat{
	ExportTexturePacker,
	ExportTexturePackerZip,
	ExportAseprite,
	ExportGodot,
	ExportGodotZip,
	ExportGrid,
	ExportGameMaker,
	ExportSequenceZip,
}

func waitGeneration(ctx context.Context, client *Client, id string) *Generation {
	for i := 0; i < 120; i++ {
		generation, err := client.GetGeneration(ctx, id, true)
		if err != nil {
			panic(err)
		}
		if generation.Status != "queued" && generation.Status != "running" {
			return generation
		}
		time.Sleep(2 * time.Second)
	}
	panic("timed out waiting for sprite generation " + id)
}

func waitSound(ctx context.Context, client *Client, id string) *SoundGeneration {
	for i := 0; i < 120; i++ {
		generation, err := client.GetSoundGeneration(ctx, id, true)
		if err != nil {
			panic(err)
		}
		if generation.Status != "queued" && generation.Status != "running" {
			return generation
		}
		time.Sleep(2 * time.Second)
	}
	panic("timed out waiting for sound generation " + id)
}

func waitAnimation(ctx context.Context, client *Client, id string) *AnimationRun {
	for i := 0; i < 180; i++ {
		run, err := client.GetAnimationRun(ctx, id)
		if err != nil {
			panic(err)
		}
		if run.Status != "queued" && run.Status != "running" {
			return run
		}
		time.Sleep(3 * time.Second)
	}
	panic("timed out waiting for animation run " + id)
}

func waitForFrames(ctx context.Context, client *Client, id string) *AnimationRun {
	lastLen := 0
	for i := 0; i < 180; i++ {
		run, err := client.GetAnimationRun(ctx, id)
		if err != nil {
			panic(err)
		}
		settled := true
		for _, frameRun := range run.FrameRuns {
			if frameRun.Status == "queued" || frameRun.Status == "running" {
				settled = false
				break
			}
		}
		if len(run.Frames) > 0 && settled && len(run.Frames) == lastLen {
			return run
		}
		lastLen = len(run.Frames)
		time.Sleep(2 * time.Second)
	}
	run, err := client.GetAnimationRun(ctx, id)
	if err != nil {
		panic(err)
	}
	return run
}

func TestLiveSpriteGenerationFullFlow(t *testing.T) {
	liveSpendEnabled(t)
	client := liveClient(t)
	ctx := context.Background()

	project, ok := createLiveProject(ctx, t, client, "Sprite Spend Test")
	if !ok {
		return
	}

	models, err := client.SpriteModels(ctx)
	if err != nil {
		t.Fatalf("sprite models: %v", err)
	}
	imageModel := ""
	for _, model := range models.ImageModels {
		if model.Available {
			imageModel = model.ID
			break
		}
	}
	if imageModel == "" {
		t.Fatal("no available image models")
	}

	job, err := client.GenerateSprite(project.ID).
		WithPrompt("a friendly red fox, side view, game sprite").
		WithMode(SpriteModeSingle).
		WithImageModel(imageModel).
		Send(ctx)
	if err != nil {
		t.Fatalf("generate sprite: %v", err)
	}

	generation := waitGeneration(ctx, client, job.ID)
	if generation.Status != "succeeded" {
		t.Fatalf("generation failed: status=%s error=%v", generation.Status, generation.Error)
	}
	if len(generation.Assets) == 0 {
		t.Fatal("no assets")
	}
	asset := generation.Assets[0]

	named, err := client.RenameAsset(ctx, asset.ID, Ptr("Hero Fox"))
	if err != nil {
		t.Fatalf("rename asset: %v", err)
	}
	if named.Name == nil || *named.Name != "Hero Fox" {
		t.Errorf("name = %v", named.Name)
	}

	metadata := map[string]string{"filepath": spriteFilepath}
	updated, err := client.PutAssetMetadata(ctx, asset.ID, metadata)
	if err != nil {
		t.Fatalf("add metadata: %v", err)
	}
	if updated.Metadata["filepath"] != spriteFilepath {
		t.Errorf("metadata = %v", updated.Metadata)
	}
	cleared, err := client.PutAssetMetadata(ctx, asset.ID, map[string]string{})
	if err != nil {
		t.Fatalf("remove metadata: %v", err)
	}
	if len(cleared.Metadata) != 0 {
		t.Errorf("metadata after clear = %v", cleared.Metadata)
	}
	if _, err := client.PutAssetMetadata(ctx, asset.ID, metadata); err != nil {
		t.Fatalf("re-add metadata: %v", err)
	}

	label, err := client.CreateLabel(ctx, project.ID, shortName("sprite"), nil)
	if err != nil {
		t.Fatalf("create label: %v", err)
	}
	associated, err := client.AssociateAssetLabel(ctx, asset.ID, label.Name)
	if err != nil {
		t.Fatalf("associate label: %v", err)
	}
	if !contains(associated.Labels, label.Name) {
		t.Errorf("labels = %v", associated.Labels)
	}
	removed, err := client.RemoveAssetLabel(ctx, asset.ID, label.Name)
	if err != nil {
		t.Fatalf("remove label: %v", err)
	}
	if contains(removed.Labels, label.Name) {
		t.Errorf("labels after remove = %v", removed.Labels)
	}
	if _, err := client.AssociateAssetLabel(ctx, asset.ID, label.Name); err != nil {
		t.Fatalf("re-associate label: %v", err)
	}

	archived, err := client.ArchiveAsset(ctx, asset.ID)
	if err != nil {
		t.Fatalf("archive: %v", err)
	}
	if archived.ArchivedAt == nil {
		t.Error("archived_at is nil")
	}
	unarchived, err := client.UnarchiveAsset(ctx, asset.ID)
	if err != nil {
		t.Fatalf("unarchive: %v", err)
	}
	if unarchived.ArchivedAt != nil {
		t.Error("archived_at should be nil after unarchive")
	}

	content, err := client.AssetContent(ctx, asset.ID)
	if err != nil {
		t.Fatalf("asset content: %v", err)
	}
	if content.IsEmpty() {
		t.Error("asset content is empty")
	}

	finishLiveProject(ctx, t, client, project.Slug)
}

func TestLiveSoundGenerationFullFlow(t *testing.T) {
	liveSpendEnabled(t)
	client := liveClient(t)
	ctx := context.Background()

	project, ok := createLiveProject(ctx, t, client, "Sound Spend Test")
	if !ok {
		return
	}

	models, err := client.SoundModels(ctx)
	if err != nil {
		t.Fatalf("sound models: %v", err)
	}
	job, err := client.GenerateSound(project.ID).
		WithPrompt("a single sword unsheathing, then a heavy metal thud").
		WithSoundModel(models.Model.ID).
		WithResponseFormat(models.DefaultFormat).
		Send(ctx)
	if err != nil {
		t.Fatalf("generate sound: %v", err)
	}

	generation := waitSound(ctx, client, job.ID)
	if generation.Status != "succeeded" {
		t.Fatalf("generation failed: status=%s error=%v", generation.Status, generation.Error)
	}
	if len(generation.Assets) == 0 {
		t.Fatal("no assets")
	}
	asset := generation.Assets[0]

	named, err := client.RenameSoundAsset(ctx, asset.ID, Ptr("Sword Unsheath"))
	if err != nil {
		t.Fatalf("rename sound: %v", err)
	}
	if named.Name == nil || *named.Name != "Sword Unsheath" {
		t.Errorf("name = %v", named.Name)
	}

	metadata := map[string]string{"filepath": soundFilepath}
	updated, err := client.PutSoundAssetMetadata(ctx, asset.ID, metadata)
	if err != nil {
		t.Fatalf("add metadata: %v", err)
	}
	if updated.Metadata["filepath"] != soundFilepath {
		t.Errorf("metadata = %v", updated.Metadata)
	}
	if _, err := client.PutSoundAssetMetadata(ctx, asset.ID, map[string]string{}); err != nil {
		t.Fatalf("remove metadata: %v", err)
	}
	if _, err := client.PutSoundAssetMetadata(ctx, asset.ID, metadata); err != nil {
		t.Fatalf("re-add metadata: %v", err)
	}

	label, err := client.CreateLabel(ctx, project.ID, shortName("sound"), nil)
	if err != nil {
		t.Fatalf("create label: %v", err)
	}
	associated, err := client.AssociateSoundLabel(ctx, asset.ID, label.Name)
	if err != nil {
		t.Fatalf("associate label: %v", err)
	}
	if !contains(associated.Labels, label.Name) {
		t.Errorf("labels = %v", associated.Labels)
	}
	if _, err := client.RemoveSoundLabel(ctx, asset.ID, label.Name); err != nil {
		t.Fatalf("remove label: %v", err)
	}
	if _, err := client.AssociateSoundLabel(ctx, asset.ID, label.Name); err != nil {
		t.Fatalf("re-associate label: %v", err)
	}

	archived, err := client.ArchiveSoundAsset(ctx, asset.ID)
	if err != nil {
		t.Fatalf("archive sound asset: %v", err)
	}
	if archived.ArchivedAt == nil {
		t.Error("archived_at is nil")
	}
	unarchived, err := client.UnarchiveSoundAsset(ctx, asset.ID)
	if err != nil {
		t.Fatalf("unarchive sound asset: %v", err)
	}
	if unarchived.ArchivedAt != nil {
		t.Error("archived_at should be nil after unarchive")
	}

	content, err := client.SoundAssetContent(ctx, asset.ID)
	if err != nil {
		t.Fatalf("sound content: %v", err)
	}
	if content.IsEmpty() {
		t.Error("sound content is empty")
	}

	finishLiveProject(ctx, t, client, project.Slug)
}

func TestLiveAnimationGenerationExportsAndSavedAnimation(t *testing.T) {
	liveSpendEnabled(t)
	client := liveClient(t)
	ctx := context.Background()

	project, ok := createLiveProject(ctx, t, client, "Animation Spend Test")
	if !ok {
		return
	}

	job, err := client.GenerateAnimation(project.ID).
		WithPrompt("the hero draws her sword and raises it overhead").
		WithAnimationModel("ash").
		WithDuration(4).
		Send(ctx)
	if err != nil {
		t.Fatalf("generate animation: %v", err)
	}

	run := waitAnimation(ctx, client, job.ID)
	if run.Status != "succeeded" {
		t.Fatalf("animation failed: status=%s error=%v", run.Status, run.Error)
	}

	if len(run.Frames) == 0 {
		if _, err := client.GenerateFrames(project.ID, run.ID).WithFPS(12).Send(ctx); err != nil {
			t.Fatalf("generate frames: %v", err)
		}
	}
	run = waitForFrames(ctx, client, run.ID)
	if len(run.Frames) == 0 {
		t.Fatal("animation produced no frames")
	}

	last := run.Frames[len(run.Frames)-1].FrameNumber
	startFrame := int64(2)
	if last < startFrame {
		startFrame = last
	}
	endFrame := last - 1
	if endFrame < startFrame {
		endFrame = startFrame
	}
	saved, err := client.SaveAnimation(project.ID).
		WithGenerationID(run.ID).
		WithRange(startFrame, endFrame).
		WithName("Sword Raise").
		Send(ctx)
	if err != nil {
		t.Fatalf("save animation: %v", err)
	}
	if saved.GenerationID != run.ID || saved.StartFrame != startFrame || saved.EndFrame != endFrame {
		t.Errorf("saved = %+v", saved)
	}

	renamed, err := client.RenameSavedAnimation(ctx, saved.ID, Ptr("Sword Raise (named)"))
	if err != nil {
		t.Fatalf("rename saved animation: %v", err)
	}
	if renamed.Name == nil || *renamed.Name != "Sword Raise (named)" {
		t.Errorf("name = %v", renamed.Name)
	}

	metadata := map[string]string{"filepath": animationFilepath}
	tagged, err := client.PutSavedAnimationMetadata(ctx, saved.ID, metadata)
	if err != nil {
		t.Fatalf("add metadata: %v", err)
	}
	if tagged.Metadata["filepath"] != animationFilepath {
		t.Errorf("metadata = %v", tagged.Metadata)
	}
	if _, err := client.PutSavedAnimationMetadata(ctx, saved.ID, map[string]string{}); err != nil {
		t.Fatalf("remove metadata: %v", err)
	}
	if _, err := client.PutSavedAnimationMetadata(ctx, saved.ID, metadata); err != nil {
		t.Fatalf("re-add metadata: %v", err)
	}

	label, err := client.CreateLabel(ctx, project.ID, shortName("anim"), nil)
	if err != nil {
		t.Fatalf("create label: %v", err)
	}
	associated, err := client.AssociateSavedAnimationLabel(ctx, saved.ID, label.Name)
	if err != nil {
		t.Fatalf("associate label: %v", err)
	}
	if !contains(associated.Labels, label.Name) {
		t.Errorf("labels = %v", associated.Labels)
	}
	if _, err := client.RemoveSavedAnimationLabel(ctx, saved.ID, label.Name); err != nil {
		t.Fatalf("remove label: %v", err)
	}
	if _, err := client.AssociateSavedAnimationLabel(ctx, saved.ID, label.Name); err != nil {
		t.Fatalf("re-associate label: %v", err)
	}

	archived, err := client.ArchiveSavedAnimation(ctx, saved.ID)
	if err != nil {
		t.Fatalf("archive saved animation: %v", err)
	}
	if archived.ArchivedAt == nil {
		t.Error("archived_at is nil")
	}
	unarchived, err := client.UnarchiveSavedAnimation(ctx, saved.ID)
	if err != nil {
		t.Fatalf("unarchive saved animation: %v", err)
	}
	if unarchived.ArchivedAt != nil {
		t.Error("archived_at should be nil after unarchive")
	}

	frame, err := client.FrameContentByNumber(ctx, run.ID, startFrame)
	if err != nil {
		t.Fatalf("frame content: %v", err)
	}
	if frame.IsEmpty() {
		t.Error("frame content is empty")
	}

	for _, format := range allFormats {
		export, err := client.Export(ctx, run.ID, format, int32(startFrame), int32(endFrame))
		if err != nil {
			t.Fatalf("export %s: %v", format, err)
		}
		switch {
		case export.Binary != nil:
			if export.Binary.IsEmpty() {
				t.Errorf("%s was empty", format)
			}
		case export.TexturePacker != nil:
			if export.TexturePacker.ImageBase64 == "" || export.TexturePacker.JSONFilename == "" {
				t.Errorf("texturepacker export incomplete")
			}
		case export.Godot != nil:
			if export.Godot.Tres == "" || export.Godot.ImageBase64 == "" {
				t.Errorf("godot export incomplete")
			}
		default:
			t.Errorf("%s returned no payload", format)
		}
	}

	finishLiveProject(ctx, t, client, project.Slug)
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
