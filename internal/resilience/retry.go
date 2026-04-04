package resilience

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

func WithRetry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error
	for i := 0; i < cfg.MaxAttempts; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		if i == cfg.MaxAttempts-1 {
			break
		}

		delay := cfg.BaseDelay * time.Duration(math.Pow(2, float64(i)))
		delay += time.Duration(rand.Intn(500)) * time.Millisecond // Jitter
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// continue
		}
	}
	return fmt.Errorf("operation failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
}