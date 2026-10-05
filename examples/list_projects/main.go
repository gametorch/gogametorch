// Command list_projects lists the projects in the caller's scope.
//
//	GAMETORCH_API_KEY=gt2_... go run ./examples/list_projects
//	GAMETORCH_API_KEY=gt2_... GAMETORCH_BASE_URL=http://localhost:8300/api \
//	  go run ./examples/list_projects
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gametorch/gogametorch"
)

func main() {
	client, err := gametorch.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	projects, err := client.ListProjects(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, project := range projects.Projects {
		fmt.Printf("%s  %s  (%s)\n", project.ID, project.Name, project.Slug)
	}
}
