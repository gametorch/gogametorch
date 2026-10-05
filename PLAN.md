# GameTorch Go SDK — implementation plan

Port the reference Rust SDK (`~/gametorch-rs`) to an idiomatic, dependency-free
Go SDK. The Rust SDK is the source of truth for behavior, wire formats, rate
limiting and error semantics.

## Constraints

- **Standard library only.** No third-party modules. In particular, no external
  UUID or decimal packages: implement a small `UUID` v4 generator with
  `crypto/rand` and a `Decimal` type on `math/big`.
- Latest stable Go (1.27.x), idiomatic module/package layout, `gofmt` + `go vet`
  clean.
- MIT licensed.
- Follow the Rust SDK's respectful client-side rate limiting, retries and
  pagination.
- Do **not** implement the CLI.

## Layout (single package `gametorch` at the module root)

| File | Contents |
| --- | --- |
| `doc.go` | package documentation |
| `client.go` | `Client`, request execution, retries, error parsing |
| `options.go` | functional options for `NewClient` |
| `error.go` | `Error` type, kinds, status predicates |
| `ratelimit.go` | per-route token buckets, shared write bucket, semaphores |
| `decimal.go` | `Decimal` (big.Int coefficient + scale), JSON (de)serialization |
| `uuid.go` | UUID v4 generation + validation helpers |
| `types.go` | shared envelopes, pagination, enums, `Download` |
| `json_helpers.go` | lenient decoders (bool-or-int, string-list, decimal-map) |
| `models_*.go` | resource structs |
| `api_*.go` | endpoint methods grouped by resource |
| `*_test.go` | offline unit, fixture and `httptest` tests |
| `live_test.go` | read-only live tests (env gated) |
| `live_writes_test.go` | non-spending live write tests (env gated) |
| `live_spend_test.go` | spending live tests (env gated) |
| `examples/` | runnable examples |

## Behavior to mirror exactly

- Default base URL `https://gametorch.app/api`; env `GAMETORCH_API_KEY`,
  `GAMETORCH_BASE_URL`.
- `Authorization: Bearer <token>`; `Accept: application/json`;
  `User-Agent: gametorch-go/<version>`.
- Rate classes: Tier1 = 1 req/s per route, Tier2 = 2 req/s per route, Writes =
  shared bucket (100 burst, 5/s), Unlimited (health).
- Concurrency: 25 outstanding holds, 5 frame generations.
- Retries: 408/429/5xx and transport errors, exponential backoff (base 500ms,
  capped 30s) with jitter, honoring numeric `Retry-After`.
- Errors: `{"error": "..."}` body, `x-request-id`/`request-id` correlation id,
  status helpers.
- Cursor pagination with `before`; `include_archived`; animation `base_asset_id`
  filter.
- Decimals: parse string or number; serialize as string; `100 credits = $1`.

## Coverage checklist (mirrors the Rust API table)

- [x] Catalog: sprite/sound/animation models
- [x] Projects: list/create/rename/delete
- [x] Sprites: generate, list/get generations, assets, content/original,
      rename, metadata, archive/unarchive/delete, generation archive/delete,
      stream
- [x] Sounds: generate, list/get, content, rename, metadata,
      archive/unarchive/delete, generation archive/unarchive, stream
- [x] Animations: estimate, generate, list/get, content,
      archive/unarchive/delete, frames, frame content/by-number, stream
- [x] Exports: plan + all eight formats
- [x] Saved animations: list/save/get/rename/metadata/archive/delete, stream
- [x] Labels: list/create/update/delete/items/thumbnail, associate/remove/
      dismiss for assets, sounds and saved animations
- [x] Art styles: list/create/generate/delete
- [x] Usage: usage, histogram, stream
- [x] API keys: list/create/update/delete
- [x] Account: ensure user, health

## Testing

- Offline: decimal round-trips, rate-limiter buckets, retry/backoff with
  `httptest`, error parsing, and JSON fixtures copied from
  `tests/serde_fixtures.rs`.
- Live: run against `http://localhost:8300/api` with `GAMETORCH_API_KEY`.
  Read and write suites are opt-in via `GAMETORCH_LIVE_WRITES=1`; spending
  suite via `GAMETORCH_LIVE_SPEND=1`.
- Finish with `gofmt -l .`, `go vet ./...`, `go test ./...` and the live suites.
