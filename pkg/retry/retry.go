package retry

import (
	"context"
	"math/rand/v2"
	"time"

	cdcerrors "github.com/foden/cdc/pkg/errors"
)

// Config holds retry configuration.
type Config struct {
	MaxAttempts int           // Maximum retry attempts (0 = infinite)
	BaseDelay   time.Duration // Initial delay between retries
	MaxDelay    time.Duration // Maximum delay cap
	Multiplier  float64       // Backoff multiplier (e.g., 2.0)
}

// DefaultConfig returns sensible defaults for transient error retry.
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
	}
}

// SourceReconnectConfig returns config for source reconnection (longer delays).
func SourceReconnectConfig() Config {
	return Config{
		MaxAttempts: 0, // infinite
		BaseDelay:   1 * time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}
}

// Do executes fn with exponential backoff + jitter.
// Stops when: fn returns nil, maxAttempts reached, ctx cancelled, or error is non-retryable.
// Non-retryable errors (wrapped with errors.Permanent) cause immediate failure without retry.
func Do(ctx context.Context, cfg Config, fn func() error) error {
	delay := cfg.BaseDelay
	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = int(^uint(0) >> 1) // effectively infinite
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Fail fast for non-retryable errors
		if cdcerrors.IsNonRetryable(lastErr) {
			return lastErr
		}

		// Don't sleep after the last attempt
		if attempt == maxAttempts-1 {
			break
		}

		// Calculate delay with jitter
		jitteredDelay := delay + jitter(delay)

		// Wait or cancel
		timer := time.NewTimer(jitteredDelay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}

		// Exponential backoff
		delay = nextDelay(delay, cfg.MaxDelay, cfg.Multiplier)
	}

	return lastErr
}

// jitter adds random delay (0 to 50% of base delay) to prevent thundering herd.
func jitter(delay time.Duration) time.Duration {
	if delay <= 0 {
		return 0
	}
	return time.Duration(rand.Int64N(int64(delay) / 2))
}

// nextDelay calculates the next backoff delay, capped at maxDelay.
func nextDelay(current, max time.Duration, multiplier float64) time.Duration {
	if multiplier <= 0 {
		multiplier = 2.0
	}
	next := time.Duration(float64(current) * multiplier)
	if max > 0 && next > max {
		return max
	}
	return next
}
