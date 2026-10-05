// Command export_animation exports an animation run in every supported format,
// writing the results to ./gametorch-exports/.
//
// This example is read-only and does not spend credits, but the run must
// already exist. Set GAMETORCH_RUN to a run id, or it picks the first run with
// frames it can find.
//
//	GAMETORCH_API_KEY=gt2_... GAMETORCH_BASE_URL=http://localhost:8300/api \
//	  go run ./examples/export_animation
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gametorch/gogametorch"
)

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

	outDir := "gametorch-exports"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatal(err)
	}

	runID := os.Getenv("GAMETORCH_RUN")
	if runID == "" {
		runID, err = findRunWithFrames(ctx, client)
		if err != nil {
			log.Fatal(err)
		}
	}
	run, err := client.GetAnimationRun(ctx, runID)
	if err != nil {
		log.Fatal(err)
	}
	last := int64(1)
	if len(run.Frames) > 0 {
		last = run.Frames[len(run.Frames)-1].FrameNumber
	}
	end := int32(3)
	if last < 3 {
		end = int32(last)
	}

	fmt.Printf("exporting run %s, frames 1-%d to %s\n", runID, end, outDir)
	for _, format := range allFormats {
		export, err := client.Export(ctx, runID, format, 1, end)
		if err != nil {
			log.Fatalf("export %s: %v", format, err)
		}
		path := filepath.Join(outDir, filename(format))
		switch {
		case export.Binary != nil:
			if err := os.WriteFile(path, export.Binary.Data, 0o644); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("  %-18s -> %s (%d bytes)\n", format, path, export.Binary.Len())
		case export.TexturePacker != nil:
			if err := writeJSON(path, export.TexturePacker); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("  %-18s -> %s\n", format, path)
		case export.Godot != nil:
			if err := writeJSON(path, export.Godot); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("  %-18s -> %s\n", format, path)
		}
	}
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func filename(format gametorch.ExportFormat) string {
	switch format {
	case gametorch.ExportTexturePacker:
		return "texturepacker.json"
	case gametorch.ExportTexturePackerZip:
		return "texturepacker.zip"
	case gametorch.ExportAseprite:
		return "animation.aseprite"
	case gametorch.ExportGodot:
		return "godot.json"
	case gametorch.ExportGodotZip:
		return "godot.zip"
	case gametorch.ExportGrid:
		return "grid.png"
	case gametorch.ExportGameMaker:
		return "gamemaker.png"
	case gametorch.ExportSequenceZip:
		return "sequence.zip"
	default:
		return "export.bin"
	}
}

func findRunWithFrames(ctx context.Context, client *gametorch.Client) (string, error) {
	params := gametorch.NewListParams().WithIncludeArchived(true)
	projects, err := client.ListProjects(ctx)
	if err != nil {
		return "", err
	}
	for _, project := range projects.Projects {
		runs, err := client.ListAnimationRuns(ctx, project.ID, params)
		if err != nil {
			return "", err
		}
		for _, run := range runs.Animations {
			if len(run.Frames) > 0 {
				return run.ID, nil
			}
		}
	}
	return "", fmt.Errorf("no animation run with frames found; set GAMETORCH_RUN to a run id")
}
