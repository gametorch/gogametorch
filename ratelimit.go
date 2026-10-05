package gametorch

import (
	"context"
	"math"
	"sync"
	"time"
)

// The per-account concurrency caps published by GameTorch.
const (
	maxOutstandingHolds = 25
	maxFrameGenerations = 5
)

// rateClass is the rate-limit bucket an operation belongs to.
type rateClass int

const (
	rateTier1     rateClass = iota // 1 request/second per route
	rateTier2                      // 2 requests/second per route
	rateWrites                     // shared token bucket across unbounded writes
	rateUnlimited                  // not rate limited (for example /health)
)

// concurrency describes the concurrency permits an operation requires.
type concurrency struct {
	hold            bool
	frameGeneration bool
}

var (
	concurrencyNone  = concurrency{}
	concurrencyHold  = concurrency{hold: true}
	concurrencyFrame = concurrency{hold: true, frameGeneration: true}
)

type bucket struct {
	tokens       float64
	last         time.Time
	capacity     float64
	refillPerSec float64
}

func newBucket(capacity, refillPerSec float64) *bucket {
	return &bucket{
		tokens:       capacity,
		last:         time.Now(),
		capacity:     capacity,
		refillPerSec: refillPerSec,
	}
}

func (b *bucket) refill(now time.Time) {
	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens = math.Min(b.tokens+elapsed*b.refillPerSec, b.capacity)
		b.last = now
	}
}

// poll consumes a token if one is available and returns the duration to wait
// otherwise. A zero duration means the request may proceed immediately.
func (b *bucket) poll(now time.Time) time.Duration {
	b.refill(now)
	if b.tokens >= 1 {
		b.tokens--
		return 0
	}
	deficit := 1 - b.tokens
	return time.Duration(deficit / b.refillPerSec * float64(time.Second))
}

func bucketParams(class rateClass) (capacity, refill float64) {
	switch class {
	case rateTier1:
		return 1, 1
	case rateTier2:
		return 1, 2
	case rateWrites:
		return 100, 5
	default:
		return math.Inf(1), math.Inf(1)
	}
}

// rateLimiter mirrors GameTorch's published limits. It is safe for concurrent
// use and shared by every clone of a Client.
type rateLimiter struct {
	enabled  bool
	mu       sync.Mutex
	perRoute map[string]*bucket
	writes   *bucket
	holds    chan struct{}
	frames   chan struct{}
}

func newRateLimiter(enabled bool) *rateLimiter {
	return &rateLimiter{
		enabled:  enabled,
		perRoute: make(map[string]*bucket),
		writes:   newBucket(100, 5),
		holds:    make(chan struct{}, maxOutstandingHolds),
		frames:   make(chan struct{}, maxFrameGenerations),
	}
}

// acquire blocks until the operation's rate bucket allows another request, or
// ctx is done.
func (l *rateLimiter) acquire(ctx context.Context, class rateClass, route string) error {
	if !l.enabled || class == rateUnlimited {
		return nil
	}
	for {
		wait := func() time.Duration {
			l.mu.Lock()
			defer l.mu.Unlock()
			now := time.Now()
			if class == rateWrites {
				return l.writes.poll(now)
			}
			b := l.perRoute[route]
			if b == nil {
				capacity, refill := bucketParams(class)
				b = newBucket(capacity, refill)
				l.perRoute[route] = b
			}
			return b.poll(now)
		}()
		if wait <= 0 {
			return nil
		}
		if err := sleepCtx(ctx, wait); err != nil {
			return err
		}
	}
}

// acquireConcurrency acquires the concurrency permits required by an operation
// and returns a function that releases them.
func (l *rateLimiter) acquireConcurrency(ctx context.Context, c concurrency) (func(), error) {
	if !l.enabled {
		return func() {}, nil
	}
	held := false
	if c.hold {
		select {
		case l.holds <- struct{}{}:
			held = true
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	framed := false
	if c.frameGeneration {
		select {
		case l.frames <- struct{}{}:
			framed = true
		case <-ctx.Done():
			if held {
				<-l.holds
			}
			return nil, ctx.Err()
		}
	}
	return func() {
		if framed {
			<-l.frames
		}
		if held {
			<-l.holds
		}
	}, nil
}
