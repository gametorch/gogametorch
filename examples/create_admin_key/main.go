// Command create_admin_key creates an admin API key with a name, spend limit,
// reset cadence and expiry, then revokes it.
//
// Must be run with an admin API key (GAMETORCH_API_KEY). Set
// GAMETORCH_KEEP_KEYS=1 to keep the key.
//
//	GAMETORCH_API_KEY=gt2_admin_... go run ./examples/create_admin_key
package main

import (
	"context"
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

	created, err := client.CreateKey(ctx, gametorch.AdminCreateApiKeyRequest().
		WithName("Admin CI key").
		WithMaxSpendLimit(gametorch.MustDecimal("5000")).
		WithSpendResetCadence(gametorch.SpendResetMonthly).
		WithExpiresAt(time.Now().Add(90*24*time.Hour)))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("created admin key: name=%v scope=%s limit=%v cadence=%s expires_at=%v\n",
		created.Name, created.KeyScope, created.MaxSpendLimit, created.SpendResetCadence, created.ExpiresAt)
	// Shown only once, at creation time. Treat it like a password.
	fmt.Printf("key_full (store securely, shown once): %s\n", created.KeyFull)

	updated, err := client.UpdateKey(ctx, created.ID, gametorch.NewUpdateApiKeyRequest().
		WithName("Admin CI key (renamed)").
		WithMaxSpendLimit(gametorch.MustDecimal("7500")).
		WithExpiresAt(time.Now().Add(180*24*time.Hour)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("updated: name=%v limit=%v expires_at=%v\n", updated.Name, updated.MaxSpendLimit, updated.ExpiresAt)

	if os.Getenv("GAMETORCH_KEEP_KEYS") == "1" {
		fmt.Printf("keeping key (GAMETORCH_KEEP_KEYS=1): %s\n", created.KeyFull)
		return
	}
	if _, err := client.DeleteKey(ctx, created.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("revoked admin key")
}
