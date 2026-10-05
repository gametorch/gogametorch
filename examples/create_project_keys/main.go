// Command create_project_keys creates a project-scoped read-only key and a
// project-scoped write key, then cleans them up.
//
// Must be run with an admin API key (GAMETORCH_API_KEY). Set
// GAMETORCH_KEEP_KEYS=1 to keep the keys and project.
//
//	GAMETORCH_API_KEY=gt2_admin_... go run ./examples/create_project_keys
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

	project, err := client.CreateProject(ctx, fmt.Sprintf("SDK Keys Example %s", shortID()))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created project %q (%s)\n", project.Name, project.Slug)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	read, err := client.CreateKey(ctx, gametorch.ProjectReadCreateApiKeyRequest(project.ID).
		WithName("Read-only CI key").
		WithMaxSpendLimit(gametorch.MustDecimal("0")).
		WithSpendResetCadence(gametorch.SpendResetMonthly).
		WithExpiresAt(expiresAt))
	if err != nil {
		log.Fatal(err)
	}
	printKey("read-only", read)

	write, err := client.CreateKey(ctx, gametorch.ProjectWriteCreateApiKeyRequest(project.ID).
		WithName("Write CI key").
		WithMaxSpendLimit(gametorch.MustDecimal("500")).
		WithSpendResetCadence(gametorch.SpendResetWeekly).
		WithExpiresAt(expiresAt))
	if err != nil {
		log.Fatal(err)
	}
	printKey("write", write)

	updated, err := client.UpdateKey(ctx, write.ID, gametorch.NewUpdateApiKeyRequest().
		WithName("Write CI key (renamed)").
		WithMaxSpendLimit(gametorch.MustDecimal("750")).
		WithSpendResetCadence(gametorch.SpendResetMonthly).
		WithExpiresAt(expiresAt.Add(30*24*time.Hour)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("updated write key: name=%v limit=%v cadence=%s expires_at=%v\n",
		updated.Name, updated.MaxSpendLimit, updated.SpendResetCadence, updated.ExpiresAt)

	keys, err := client.ListKeys(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nall keys:")
	for _, key := range keys.Keys {
		name := "<unnamed>"
		if key.Name != nil {
			name = *key.Name
		}
		fmt.Printf("  %-24s scope=%-13s project=%v limit=%v cadence=%s prefix=%s\n",
			name, key.KeyScope, key.ProjectID, key.MaxSpendLimit, key.SpendResetCadence, key.KeyPrefix)
	}

	if os.Getenv("GAMETORCH_KEEP_KEYS") == "1" {
		fmt.Println("\nkeeping keys and project (GAMETORCH_KEEP_KEYS=1)")
		fmt.Printf("  read key:  %s\n", read.KeyFull)
		fmt.Printf("  write key: %s\n", write.KeyFull)
		return
	}
	if _, err := client.DeleteKey(ctx, read.ID); err != nil {
		log.Fatal(err)
	}
	if _, err := client.DeleteKey(ctx, write.ID); err != nil {
		log.Fatal(err)
	}
	if _, err := client.DeleteProject(ctx, project.Slug); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nrevoked keys and deleted project")
}

func printKey(kind string, created *gametorch.ApiKeyWithSecret) {
	fmt.Printf("created %s key: scope=%s project=%v limit=%v cadence=%s expires_at=%v\n",
		kind, created.KeyScope, created.ProjectID, created.MaxSpendLimit, created.SpendResetCadence, created.ExpiresAt)
	fmt.Printf("  key_full (store securely, shown once): %s\n", created.KeyFull)
}

func shortID() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
