# GameTorch Go SDK

The official Go SDK for the [GameTorch](https://gametorch.app) API.
GameTorch generates game-ready **sprites**, **sound effects** and **animations**
from text prompts and organizes them into projects.

- [API reference](https://gametorch.app/api/docs)
- [Agent / LLM guide](https://gametorch.app/llms.txt)
- [Privacy policy](https://gametorch.app/privacy)
- [Terms and conditions](https://gametorch.app/terms)

## Features

- Full coverage of the public GameTorch API, with strongly typed request and
  response models.
- **Standard library only** — no third-party dependencies.
- Polite by default: client-side rate limiting that mirrors GameTorch's
  published limits, concurrency caps, and automatic retries with exponential
  backoff that honors `Retry-After`.
- Cursor pagination with `Paginator` helpers.
- Accurate money handling with a `Decimal` type built on `math/big`
  (100 credits = $1).
- Sensible errors with status-code helpers.
- MIT licensed.

Requires Go 1.22 or newer.

## Installation

```sh
go get github.com/gametorch/gogametorch
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gametorch/gogametorch"
)

func main() {
	// Reads GAMETORCH_API_KEY; or pass gametorch.WithAPIKey("gt2_...").
	client, err := gametorch.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	models, err := client.SpriteModels(ctx)
	if err != nil {
		log.Fatal(err)
	}
	project, err := client.CreateProject(ctx, "My Game")
	if err != nil {
		log.Fatal(err)
	}

	job, err := client.GenerateSprite(project.ID).
		WithPrompt("a red fox, side view").
		WithMode(gametorch.SpriteModeSingle).
		WithImageModel(models.ImageModels[0].ID).
		Send(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("generation %s is %s\n", job.ID, job.Status)
}
```

## Authentication

Every request uses `Authorization: Bearer <token>`, where the token is either a
server-to-server API key (`gt2_...`) or a Clerk session token:

```go
// API key (server-to-server).
client, err := gametorch.NewClient(gametorch.WithAPIKey("gt2_..."))

// Clerk session token (browser / trusted backend).
client, err := gametorch.NewClient(gametorch.WithBearerToken("eyJ..."))
```

Create an API key from the GameTorch dashboard or with `Client.CreateKey`. Keys
are shown only once and can carry a spend limit.

### Key scopes

Every API key has a `key_scope` fixed at creation:

- `admin` — full access to the owning account/organization, exactly like an
  admin session. Carries no project.
- `project_write` — bound to one project: read and write its sprites, sounds,
  animations, labels, names and metadata. No account/admin access, no other
  projects.
- `project_read` — bound to one project: read-only.

Project-scoped keys get `403` on account/admin routes and `404` when they name
another project. Build scoped keys with the request constructors:

```go
read, err := client.CreateKey(ctx, gametorch.ProjectReadCreateApiKeyRequest(project.ID).
	WithName("CI read key").
	WithMaxSpendLimit(gametorch.MustDecimal("0")).
	WithSpendResetCadence(gametorch.SpendResetMonthly).
	WithExpiresAt(time.Now().Add(30*24*time.Hour)))
fmt.Printf("scope: %s, project: %v\n", read.KeyScope, read.ProjectID)
```

See `ApiKeyScope`, `CreateApiKeyRequest` and the `create_admin_key` /
`create_project_keys` examples.

## Base URL

The SDK defaults to `https://gametorch.app/api`. For local development, point it
at your local deployment:

```go
client, err := gametorch.NewClient(
	gametorch.WithAPIKey("gt2_local_dev_key"),
	gametorch.WithBaseURL("http://localhost:8300/api"),
)
```

You can also set `GAMETORCH_BASE_URL`.

## Rate limits

GameTorch rate limits every account per route and returns `429` when a limit is
exceeded. The SDK is respectful by default:

- Throttles **Tier 1** routes (generation creates, frame generation, animation
  exports) to 1 request/second per route.
- Throttles **Tier 2** routes (content, usage, search, single-item reads) to
  2 requests/second per route.
- Shares a **token bucket** (100-request burst, 5 requests/second refill) across
  unbounded writes.
- Caps in-flight hold-creating requests at 25 and concurrent frame generations
  at 5.
- Retries `429`, `408` and `5xx` responses with exponential backoff and jitter,
  honoring `Retry-After`.

Tune or disable this behavior on the client:

```go
client, err := gametorch.NewClient(
	gametorch.WithAPIKey("gt2_..."),
	gametorch.WithMaxRetries(5),
	gametorch.WithRetryBaseDelay(250*time.Millisecond),
	gametorch.WithRateLimit(false), // only if you manage limits yourself
)
```

## Pagination

List endpoints accept `*ListParams` and return a page with a `NextCursor`. The
`Stream*` helpers fetch pages on demand:

```go
generations := client.StreamGenerations(project.ID, false)
for {
	generation, ok, err := generations.Next(ctx)
	if err != nil {
		return err
	}
	if !ok {
		break
	}
	fmt.Println(generation.ID)
}
```

## Filtering animations by base image

Animation runs link back to the sprite asset they were generated from via
`BaseAssetID` (`nil` when generated from scratch). You can filter animation
queries by it:

```go
params := gametorch.NewListParams().WithBaseAssetID(spriteID)
runs, err := client.ListAnimationRuns(ctx, project.ID, params)
for _, run := range runs.Animations {
	fmt.Println(run.ID, run.BaseAssetID)
}

// The stream helper honors the same filters.
stream := client.StreamAnimationRuns(project.ID, params)
for {
	run, ok, err := stream.Next(ctx)
	// ...
}
```

## Provenance

Generation and asset responses expose who or what created them. The fields are
flattened onto the resource as `Provenance` (`UserID`, `Source`, `APIKeyID`,
`KeyName`):

```go
generation, err := client.GetGeneration(ctx, generationID, false)
fmt.Printf("created by %s via %s\n", generation.Provenance.UserID, generation.Provenance.Source)
```

## Error handling

Every fallible operation returns `error` whose concrete type is
`*gametorch.Error`, exposing the HTTP status, the API's `error` message and
convenience predicates:

```go
_, err := client.GetAsset(ctx, assetID)
var apiErr *gametorch.Error
switch {
case err == nil:
	fmt.Println("found")
case errors.As(err, &apiErr) && apiErr.IsNotFound():
	fmt.Println("no such asset")
case errors.As(err, &apiErr) && apiErr.IsRateLimited():
	fmt.Println("slow down")
default:
	return err
}
```

Available predicates: `IsNotFound`, `IsUnauthorized`, `IsPaymentRequired`,
`IsForbidden`, `IsConflict` and `IsRateLimited`. `Error.Kind()` returns a
stable label, and `Error.APIMessage()` / `Error.RequestID()` expose the API
message and correlation id.

## API coverage

| Area | Methods |
| --- | --- |
| Catalog | `SpriteModels`, `SoundModels`, `AnimationModels` |
| Projects | `ListProjects`, `CreateProject`, `RenameProject`, `DeleteProject` |
| Sprites | `GenerateSprite`, `ListGenerations`, `GetGeneration`, `ListSpriteAssets`, `GetAsset`, `AssetContent`, `AssetOriginal`, `RenameAsset`, `PutAssetMetadata`, `ArchiveAsset`, `UnarchiveAsset`, `DeleteAsset`, `ArchiveGeneration`, `UnarchiveGeneration`, `DeleteGeneration` |
| Sounds | `GenerateSound`, `ListSoundGenerations`, `GetSoundGeneration`, `SoundAssetContent`, `RenameSoundAsset`, `PutSoundAssetMetadata`, `ArchiveSoundAsset`, `UnarchiveSoundAsset`, `DeleteSoundAsset`, `ArchiveSoundGeneration`, `UnarchiveSoundGeneration` |
| Animations | `EstimateAnimation`, `GenerateAnimation`, `ListAnimationRuns`, `GetAnimationRun`, `AnimationContent`, `ArchiveAnimationRun`, `UnarchiveAnimationRun`, `DeleteAnimationRun`, `GenerateFrames`, `FrameContent`, `FrameContentByNumber` |
| Exports | `ExportPlan`, `Export`, `ExportTexturePacker`, `ExportTexturePackerZip`, `ExportAseprite`, `ExportGodot`, `ExportGodotZip`, `ExportGrid`, `ExportGameMaker`, `ExportSequenceZip` |
| Saved animations | `ListSavedAnimations`, `SaveAnimation`, `GetSavedAnimation`, `RenameSavedAnimation`, `PutSavedAnimationMetadata`, `ArchiveSavedAnimation`, `UnarchiveSavedAnimation`, `DeleteSavedAnimation` |
| Labels | `ListLabels`, `CreateLabel`, `UpdateLabel`, `DeleteLabel`, `LabelItems`, `SetLabelThumbnail`, `AssociateAssetLabel`, `RemoveAssetLabel`, `DismissAssetLabelSuggestion`, `AssociateSoundLabel`, `RemoveSoundLabel`, `DismissSoundLabelSuggestion`, `AssociateSavedAnimationLabel`, `RemoveSavedAnimationLabel` |
| Art styles | `ListArtStyles`, `CreateArtStyle`, `GenerateArtStyle`, `DeleteArtStyle` |
| Usage | `Usage`, `UsageHistogram`, `StreamUsage` |
| API keys | `ListKeys`, `CreateKey`, `UpdateKey`, `DeleteKey` |
| Account | `EnsureUser`, `Health` |

## Examples

Runnable examples live in [`examples/`](examples):

```sh
export GAMETORCH_API_KEY=gt2_...
go run ./examples/list_projects
go run ./examples/generate_sprite      # spends credits
go run ./examples/generate_sound       # spends credits
go run ./examples/generate_animation   # spends credits
go run ./examples/export_animation     # read-only; exports every format
go run ./examples/create_admin_key     # needs an admin key
go run ./examples/create_project_keys  # needs an admin key
GAMETORCH_PROJECT=<project-id> go run ./examples/stream_generations
```

`generate_animation` walks the full animation workflow: generation, frame
generation, saving a sub-range as a named preset, naming, metadata, labels and
exporting in every supported format. `create_project_keys` creates a
project-scoped read-only key and write key with names, spend limits, reset
cadence and expiry.

## Testing

`go test ./...` runs the offline unit, fixture and `httptest` suites by
default. The live suites are opt-in so they never hit the network or spend
credits unless you ask:

| Suite | Env vars | Spends credits |
| --- | --- | --- |
| `live_test.go` | `GAMETORCH_API_KEY` | No (read-only + free estimate) |
| `live_writes_test.go` | `GAMETORCH_API_KEY`, `GAMETORCH_LIVE_WRITES=1` | No (creates and cleans up its own data) |
| `live_spend_test.go` | `GAMETORCH_API_KEY`, `GAMETORCH_LIVE_SPEND=1` | **Yes** |

```sh
GAMETORCH_API_KEY=gt2_... GAMETORCH_BASE_URL=http://localhost:8300/api \
  go test -run TestLive -v

GAMETORCH_API_KEY=gt2_... GAMETORCH_LIVE_WRITES=1 \
  go test -run TestLive -v

# Costs money: generates a sprite, a sound and a 4s animation plus exports.
GAMETORCH_API_KEY=gt2_... GAMETORCH_LIVE_SPEND=1 \
  go test -run TestLive -v -timeout 30m
```

The live tests and the generation examples each create their own project. By
default they delete it again at the end; set `GAMETORCH_KEEP_PROJECT=1` to keep
the project (and its generations, names, labels and metadata) so you can inspect
it in the GameTorch UI.

## License

MIT. See [LICENSE](LICENSE).
