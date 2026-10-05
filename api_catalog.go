package gametorch

import (
	"context"
	"net/http"
)

// SpriteModels returns the image-model catalog, prompt-enhancement text models,
// supported modes and reservation information.
//
// GET /sprite-models
func (c *Client) SpriteModels(ctx context.Context) (*SpriteModels, error) {
	out, err := fetchJSON[SpriteModels](c, ctx, rateTier2, "GET /sprite-models", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "sprite-models",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// SoundModels returns the sound-model catalog and supported output formats.
//
// GET /sound-models
func (c *Client) SoundModels(ctx context.Context) (*SoundModels, error) {
	out, err := fetchJSON[SoundModels](c, ctx, rateTier2, "GET /sound-models", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "sound-models",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AnimationModels returns the animation models by their public names (ash,
// birch, cedar), with their descriptions, durations and resolutions.
//
// GET /animation-models
func (c *Client) AnimationModels(ctx context.Context) (*AnimationModels, error) {
	out, err := fetchJSON[AnimationModels](c, ctx, rateTier2, "GET /animation-models", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "animation-models",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
