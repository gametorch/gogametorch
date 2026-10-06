// Command animate_sprite generates a sprite and then animates it by passing the
// sprite asset's id as the animation's base asset, so the animation stays
// visually consistent with the sprite.
//
// This example spends credits. It creates a fresh project and deletes it again
// at the end; set GAMETORCH_KEEP_PROJECT=1 to keep it.
//
//	GAMETORCH_API_KEY=gt2_... go run ./examples/animate_sprite
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

func main() {
	client, err := gametorch.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	project, err := client.CreateProject(ctx, projectName("Animate Sprite Example"))
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
	// 1. Generate the sprite the animation will be based on.
	sprite, err := generateSprite(ctx, client, project.ID)
	if err != nil {
		return err
	}
	fmt.Printf("sprite asset %s (%dx%d)\n", sprite.ID, sprite.Width, sprite.Height)

	// 2. Estimate and start an animation that references that sprite asset.
	estimate, err := client.EstimateAnimation(project.ID).
		WithAnimationModel("ash").
		WithDuration(4).
		WithBaseAssetID(sprite.ID).
		Send(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("estimate: %s credits (%s USD)\n", estimate.Credits, estimate.USD)

	job, err := client.GenerateAnimation(project.ID).
		WithPrompt("the hero draws her sword and raises it overhead").
		WithAnimationModel("ash").
		WithDuration(4).
		WithBaseAssetID(sprite.ID).
		Send(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("animation %s is %s\n", job.ID, job.Status)

	run, err := waitForAnimation(ctx, client, job.ID)
	if err != nil {
		return err
	}
	if run.BaseAssetID == nil || *run.BaseAssetID != sprite.ID {
		return fmt.Errorf("animation base_asset_id = %v, want %s", run.BaseAssetID, sprite.ID)
	}
	fmt.Printf("animation %s references sprite asset %s\n", run.ID, *run.BaseAssetID)

	// 3. Filter animation runs back to the sprite they were based on.
	filtered, err := client.ListAnimationRuns(ctx, project.ID, gametorch.NewListParams().WithBaseAssetID(sprite.ID))
	if err != nil {
		return err
	}
	fmt.Printf("%d animation run(s) based on %s\n", filtered.Total, sprite.ID)
	return nil
}

func generateSprite(ctx context.Context, client *gametorch.Client, projectID string) (*gametorch.Asset, error) {
	models, err := client.SpriteModels(ctx)
	if err != nil {
		return nil, err
	}
	imageModel := ""
	for _, model := range models.ImageModels {
		if model.Available {
			imageModel = model.ID
			break
		}
	}
	if imageModel == "" {
		return nil, fmt.Errorf("no available image models")
	}

	job, err := client.GenerateSprite(projectID).
		WithPrompt("a friendly red fox, side view, game sprite").
		WithMode(gametorch.SpriteModeSingle).
		WithImageModel(imageModel).
		Send(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("sprite generation %s is %s\n", job.ID, job.Status)

	for {
		generation, err := client.GetGeneration(ctx, job.ID, true)
		if err != nil {
			return nil, err
		}
		if generation.Status != "queued" && generation.Status != "running" {
			if len(generation.Assets) == 0 {
				return nil, fmt.Errorf("generation produced no assets (status: %s)", generation.Status)
			}
			return &generation.Assets[0], nil
		}
		time.Sleep(2 * time.Second)
	}
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
