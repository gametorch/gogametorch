// Command generate_animation creates a project, generates an animation and its
// frames, saves a sub-range as a named preset, and exports every format.
//
// This example spends credits. It creates a fresh project and deletes it again
// at the end; set GAMETORCH_KEEP_PROJECT=1 to keep it.
//
//	GAMETORCH_API_KEY=gt2_... go run ./examples/generate_animation
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gametorch/gogametorch"
)

const filepath = "/foo/bar/animations/sword-raise.aseprite"

var allFormats = []gametorch.ExportFormat{
	gametorch.ExportTexturePacker,
	gametorch.ExportTexturePackerZip,
	gametorch.ExportAseprite,
	gametorch.ExportGodot,
	gametorch.ExportGodotZip,
	gametorch.ExportGrid,
	gametorch.ExportGameMaker,
	gametorch.ExportSequenceZip,
}

func main() {
	client, err := gametorch.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	project, err := client.CreateProject(ctx, projectName("Animation Example"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created project %q (%s)\n", project.Name, project.Slug)

	runErr := run(ctx, client, project)
	finish(ctx, client, project.Slug)
	if runErr != nil {
		log.Fatal(runErr)
	}
}

func run(ctx context.Context, client *gametorch.Client, project *gametorch.Project) error {
	job, err := client.GenerateAnimation(project.ID).
		WithPrompt("the hero draws her sword and raises it overhead").
		WithAnimationModel("ash").
		WithDuration(4).
		Send(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("animation %s is %s\n", job.ID, job.Status)

	run, err := waitForAnimation(ctx, client, job.ID)
	if err != nil {
		return err
	}
	fmt.Printf("animation %s finished with %d frame(s)\n", run.ID, len(run.Frames))

	if len(run.Frames) == 0 {
		framesJob, err := client.GenerateFrames(project.ID, run.ID).WithFPS(12).Send(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("frame generation %s is %s\n", framesJob.ID, framesJob.Status)
	}
	run, err = waitForFrames(ctx, client, run.ID)
	if err != nil {
		return err
	}
	if len(run.Frames) == 0 {
		fmt.Println("no frames available; skipping saved animation and exports")
		return nil
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
		return err
	}
	fmt.Printf("saved animation %s (%d frames, %d-%d)\n", saved.ID, saved.FrameCount, saved.StartFrame, saved.EndFrame)

	renamed, err := client.RenameSavedAnimation(ctx, saved.ID, gametorch.Ptr("Sword Raise (named)"))
	if err != nil {
		return err
	}
	fmt.Printf("renamed saved animation: %v\n", renamed.Name)

	metadata := map[string]string{"filepath": filepath}
	tagged, err := client.PutSavedAnimationMetadata(ctx, saved.ID, metadata)
	if err != nil {
		return err
	}
	fmt.Printf("metadata after add: %v\n", tagged.Metadata)
	cleared, err := client.PutSavedAnimationMetadata(ctx, saved.ID, map[string]string{})
	if err != nil {
		return err
	}
	fmt.Printf("metadata after remove: %v\n", cleared.Metadata)
	if _, err := client.PutSavedAnimationMetadata(ctx, saved.ID, metadata); err != nil {
		return err
	}

	labelName := fmt.Sprintf("anim-%s", shortID())
	label, err := client.CreateLabel(ctx, project.ID, labelName, gametorch.Ptr("#ff9800"))
	if err != nil {
		return err
	}
	associated, err := client.AssociateSavedAnimationLabel(ctx, saved.ID, label.Name)
	if err != nil {
		return err
	}
	fmt.Printf("labels after add: %v\n", associated.Labels)
	removed, err := client.RemoveSavedAnimationLabel(ctx, saved.ID, label.Name)
	if err != nil {
		return err
	}
	fmt.Printf("labels after remove: %v\n", removed.Labels)
	if _, err := client.AssociateSavedAnimationLabel(ctx, saved.ID, label.Name); err != nil {
		return err
	}
	fmt.Printf("kept label %q\n", label.Name)

	frame, err := client.FrameContentByNumber(ctx, run.ID, startFrame)
	if err != nil {
		return err
	}
	fmt.Printf("frame %d: %d bytes\n", startFrame, frame.Len())

	fmt.Printf("\nexports (%d-%d):\n", startFrame, endFrame)
	for _, format := range allFormats {
		export, err := client.Export(ctx, run.ID, format, int32(startFrame), int32(endFrame))
		if err != nil {
			return fmt.Errorf("export %s: %w", format, err)
		}
		switch {
		case export.Binary != nil:
			fmt.Printf("  %-18s %d bytes (%s)\n", format, export.Binary.Len(), export.Binary.ContentType)
		case export.TexturePacker != nil:
			fmt.Printf("  %-18s json, image %d bytes, atlas %s\n", format, len(export.TexturePacker.ImageBase64), export.TexturePacker.JSONFilename)
		case export.Godot != nil:
			fmt.Printf("  %-18s json, image %d bytes, resource %s\n", format, len(export.Godot.ImageBase64), export.Godot.TresFilename)
		}
	}
	return nil
}

func waitForAnimation(ctx context.Context, client *gametorch.Client, runID string) (*gametorch.AnimationRun, error) {
	for {
		run, err := client.GetAnimationRun(ctx, runID)
		if err != nil {
			return nil, err
		}
		fmt.Println("status:", run.Status)
		if run.Status != "queued" && run.Status != "running" {
			return run, nil
		}
		time.Sleep(3 * time.Second)
	}
}

func waitForFrames(ctx context.Context, client *gametorch.Client, runID string) (*gametorch.AnimationRun, error) {
	lastLen := 0
	for i := 0; i < 180; i++ {
		run, err := client.GetAnimationRun(ctx, runID)
		if err != nil {
			return nil, err
		}
		settled := true
		for _, frameRun := range run.FrameRuns {
			if frameRun.Status == "queued" || frameRun.Status == "running" {
				settled = false
				break
			}
		}
		if len(run.Frames) > 0 && settled && len(run.Frames) == lastLen {
			return run, nil
		}
		lastLen = len(run.Frames)
		time.Sleep(2 * time.Second)
	}
	return client.GetAnimationRun(ctx, runID)
}

func projectName(kind string) string {
	return fmt.Sprintf("SDK %s %s", kind, shortID())
}

func shortID() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func finish(ctx context.Context, client *gametorch.Client, slug string) {
	if os.Getenv("GAMETORCH_KEEP_PROJECT") == "1" {
		fmt.Printf("keeping project %q (GAMETORCH_KEEP_PROJECT=1)\n", slug)
		return
	}
	if _, err := client.DeleteProject(ctx, slug); err != nil {
		log.Printf("delete project: %v", err)
		return
	}
	fmt.Printf("deleted project %q\n", slug)
}
