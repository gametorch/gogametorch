// Package gametorch is the official Go SDK for the GameTorch API. GameTorch
// generates game-ready sprites, sound effects and animations from text prompts
// and organizes them into projects.
//
// # Getting started
//
//	client, err := gametorch.NewClientFromEnv()
//	if err != nil {
//		return err
//	}
//
//	models, err := client.SpriteModels(ctx)
//	project, err := client.CreateProject(ctx, "My Game")
//	job, err := client.GenerateSprite(project.ID).
//		WithPrompt("a red fox, side view").
//		WithImageModel(models.ImageModels[0].ID).
//		Send(ctx)
//
// # Authentication
//
// Every request is authenticated with "Authorization: Bearer <token>", where
// the token is either a server-to-server API key (gt2_...) or a Clerk session
// token. Use WithAPIKey or WithBearerToken, or set GAMETORCH_API_KEY and call
// NewClientFromEnv.
//
// # Base URL
//
// The client defaults to DefaultBaseURL (https://gametorch.app/api). Point it
// at a local deployment with WithBaseURL("http://localhost:8300/api"), or set
// GAMETORCH_BASE_URL.
//
// # Rate limits
//
// GameTorch rate limits each account per route. The SDK is polite by default:
// it throttles requests locally to match GameTorch's published limits, caps
// concurrent hold-creating requests and frame generations, and automatically
// retries 429/5xx responses with exponential backoff (honoring Retry-After).
// This is configurable with the options on NewClient; disable it with
// WithRateLimit(false) if you manage limits yourself.
//
// # Pagination
//
// List endpoints take *ListParams and return a page with a next cursor. The
// Stream* methods return a *Paginator that fetches pages on demand.
//
// # Errors
//
// All fallible operations return *Error, which exposes the HTTP status and API
// message plus helpers such as IsNotFound and IsRateLimited.
//
// Documentation:
//
//   - API reference: https://gametorch.app/api/docs
//   - Agent / LLM guide: https://gametorch.app/llms.txt
//   - Privacy policy: https://gametorch.app/privacy
//   - Terms and conditions: https://gametorch.app/terms
package gametorch
