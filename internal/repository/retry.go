package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/iamgideonidoko/signet/pkg/logger"
)

var (
	ErrNoConnection = errors.New("no database connection")
	ErrMaxRetries   = errors.New("max retries exceeded")
)

type RetryConfig struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 5,
	InitialWait: 100 * time.Millisecond,
	MaxWait:     5 * time.Second,
	Multiplier:  2.0,
}

// WithRetry wraps a database operation with exponential backoff retry logic.
func WithRetry(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error
	wait := config.InitialWait

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on context cancellation
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		// Don't retry on certain SQL errors
		if errors.Is(err, sql.ErrNoRows) {
			return err
		}

		// Check if we've exhausted retries
		if attempt >= config.MaxAttempts {
			break
		}

		// Log retry attempt
		logger.Warn("Database operation failed, retrying", map[string]any{
			"attempt": attempt,
			"max":     config.MaxAttempts,
			"wait_ms": wait.Milliseconds(),
			"error":   err.Error(),
		})

		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}

		// Exponential backoff
		wait = time.Duration(float64(wait) * config.Multiplier)
		wait = min(wait, config.MaxWait)
	}

	return fmt.Errorf("%w: %v", ErrMaxRetries, lastErr)
}

// HealthCheck verifies database connectivity.
func (r *Repository) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.db.PingContext(ctx)
}

// Stats returns database connection pool statistics.
func (r *Repository) Stats() sql.DBStats {
	return r.db.Stats()
}
