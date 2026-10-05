// Command stream_generations streams every generation in a project, page by
// page.
//
//	GAMETORCH_API_KEY=gt2_... GAMETORCH_PROJECT=<project-id> \
//	  go run ./examples/stream_generations
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gametorch/gogametorch"
)

func main() {
	client, err := gametorch.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	projectID := os.Getenv("GAMETORCH_PROJECT")
	if projectID == "" {
		log.Fatal("set GAMETORCH_PROJECT to a project id")
	}

	ctx := context.Background()
	generations := client.StreamGenerations(projectID, false)
	count := 0
	for {
		generation, ok, err := generations.Next(ctx)
		if err != nil {
			log.Fatal(err)
		}
		if !ok {
			break
		}
		count++
		fmt.Printf("%s  %s  %d asset(s)\n", generation.ID, generation.Status, len(generation.Assets))
	}
	total, _ := generations.Total()
	fmt.Printf("streamed %d generation(s), total=%d\n", count, total)
}
