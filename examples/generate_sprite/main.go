// Command generate_sprite creates a project, generates a sprite, then names it
// and tags it with metadata and a label.
//
// This example spends credits. It creates a fresh project and deletes it again
// at the end; set GAMETORCH_KEEP_PROJECT=1 to keep it.
//
//	GAMETORCH_API_KEY=gt2_... go run ./examples/generate_sprite
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

const filepath = "/foo/bar/sprites/hero.png"

func main() {
	client, err := gametorch.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	project, err := client.CreateProject(ctx, projectName("Sprite Example"))
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
	models, err := client.SpriteModels(ctx)
	if err != nil {
		return err
	}
	imageModel := ""
	for _, model := range models.ImageModels {
		if model.Available {
			imageModel = model.ID
			break
		}
	}
	if imageModel == "" {
		return fmt.Errorf("no available image models")
	}

	job, err := client.GenerateSprite(project.ID).
		WithPrompt("a friendly red fox, side view, game sprite").
		WithMode(gametorch.SpriteModeSingle).
		WithImageModel(imageModel).
		Send(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("generation %s is %s\n", job.ID, job.Status)

	var generation *gametorch.Generation
	for {
		generation, err = client.GetGeneration(ctx, job.ID, true)
		if err != nil {
			return err
		}
		fmt.Println("status:", generation.Status)
		if generation.Status != "queued" && generation.Status != "running" {
			break
		}
		time.Sleep(2 * time.Second)
	}

	if len(generation.Assets) == 0 {
		fmt.Printf("generation produced no assets (status: %s)\n", generation.Status)
		return nil
	}
	asset := generation.Assets[0]
	fmt.Printf("asset %s (%dx%d)\n", asset.ID, asset.Width, asset.Height)

	named, err := client.RenameAsset(ctx, asset.ID, gametorch.Ptr("Hero Fox"))
	if err != nil {
		return err
	}
	fmt.Printf("named asset: %v\n", named.Name)

	metadata := map[string]string{"filepath": filepath}
	updated, err := client.PutAssetMetadata(ctx, asset.ID, metadata)
	if err != nil {
		return err
	}
	fmt.Printf("metadata after add: %v\n", updated.Metadata)
	cleared, err := client.PutAssetMetadata(ctx, asset.ID, map[string]string{})
	if err != nil {
		return err
	}
	fmt.Printf("metadata after remove: %v\n", cleared.Metadata)
	if _, err := client.PutAssetMetadata(ctx, asset.ID, metadata); err != nil {
		return err
	}

	labelName := fmt.Sprintf("sprite-%s", shortID())
	label, err := client.CreateLabel(ctx, project.ID, labelName, gametorch.Ptr("#4caf50"))
	if err != nil {
		return err
	}
	associated, err := client.AssociateAssetLabel(ctx, asset.ID, label.Name)
	if err != nil {
		return err
	}
	fmt.Printf("labels after add: %v\n", associated.Labels)
	removed, err := client.RemoveAssetLabel(ctx, asset.ID, label.Name)
	if err != nil {
		return err
	}
	fmt.Printf("labels after remove: %v\n", removed.Labels)
	if _, err := client.AssociateAssetLabel(ctx, asset.ID, label.Name); err != nil {
		return err
	}
	fmt.Printf("kept label %q\n", label.Name)

	png, err := client.AssetContent(ctx, asset.ID)
	if err != nil {
		return err
	}
	fmt.Printf("downloaded %d bytes of %s\n", png.Len(), png.ContentType)
	return nil
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
