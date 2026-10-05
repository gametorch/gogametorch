package gametorch

import (
	"context"
	"net/http"
	"net/url"
)

// Usage returns the credit balance (admins), a per-source summary and the
// operation log.
//
//   - before is the pagination cursor.
//   - split may be "user" to break usage out per member.
//
// GET /usage
func (c *Client) Usage(ctx context.Context, before, split *string) (*Usage, error) {
	query := url.Values{}
	if before != nil {
		query.Set("before", *before)
	}
	if split != nil {
		query.Set("split", *split)
	}
	out, err := fetchJSON[Usage](c, ctx, rateTier2, "GET /usage", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "usage",
		query:  query,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UsageHistogram returns dense time buckets of spend split by source.
//
//   - range is "24h", "7d", "30d" or "365d".
//   - source filters to a single spend source.
//   - split may be "user" to break usage out per member.
//
// GET /usage/histogram
func (c *Client) UsageHistogram(ctx context.Context, rangeValue, source, split *string) (*UsageHistogram, error) {
	query := url.Values{}
	if rangeValue != nil {
		query.Set("range", *rangeValue)
	}
	if source != nil {
		query.Set("source", *source)
	}
	if split != nil {
		query.Set("split", *split)
	}
	out, err := fetchJSON[UsageHistogram](c, ctx, rateTier2, "GET /usage/histogram", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "usage/histogram",
		query:  query,
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// StreamUsage streams the operation log page by page.
func (c *Client) StreamUsage(split *string) *Paginator[UsageRecord] {
	return newPaginator(func(ctx context.Context, cursor *string) (Page[UsageRecord], error) {
		response, err := c.Usage(ctx, cursor, split)
		if err != nil {
			return Page[UsageRecord]{}, err
		}
		return Page[UsageRecord]{
			Items:      response.Records,
			NextCursor: response.NextCursor,
			Total:      response.Total,
		}, nil
	})
}
