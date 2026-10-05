package gametorch

import (
	"context"
	"testing"
	"time"
)

func TestBucketAllowsBurstThenThrottles(t *testing.T) {
	b := newBucket(2, 1)
	now := time.Now()
	if wait := b.poll(now); wait != 0 {
		t.Fatalf("first poll should not wait, got %v", wait)
	}
	if wait := b.poll(now); wait != 0 {
		t.Fatalf("second poll should not wait, got %v", wait)
	}
	if wait := b.poll(now); wait < 900*time.Millisecond {
		t.Fatalf("third poll should throttle ~1s, got %v", wait)
	}
}

func TestWritesBucketHasLargeBurst(t *testing.T) {
	b := newBucket(100, 5)
	now := time.Now()
	for i := 0; i < 100; i++ {
		if wait := b.poll(now); wait != 0 {
			t.Fatalf("poll %d should not wait, got %v", i, wait)
		}
	}
	if wait := b.poll(now); wait == 0 {
		t.Fatal("101st poll should throttle")
	}
}

func TestDisabledLimiterNeverBlocks(t *testing.T) {
	limiter := newRateLimiter(false)
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		if err := limiter.acquire(ctx, rateTier1, "route"); err != nil {
			t.Fatal(err)
		}
	}
	release, err := limiter.acquireConcurrency(ctx, concurrencyFrame)
	if err != nil {
		t.Fatal(err)
	}
	release()
}

func TestLimiterAcquireHonorsContext(t *testing.T) {
	limiter := newRateLimiter(true)
	ctx := context.Background()
	// Exhaust the Tier1 bucket, then cancel.
	if err := limiter.acquire(ctx, rateTier1, "ctx-route"); err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := limiter.acquire(cancelled, rateTier1, "ctx-route"); err == nil {
		t.Fatal("expected context cancellation error")
	}
}
