package storage

import (
	"context"
	"time"
)

// Result reports the outcome of a rate-limit check.
type Result struct {
	Allowed   bool
	Remaining int64
	ResetAt   time.Time
	Limit     int64
}

// RateLimitStorage is the interface for rate-limit backends (e.g., Redis, memory, SQL).
type RateLimitStorage interface {
	CheckAndUpdate(ctx context.Context, key string, limit int64, window time.Duration, cost int64) (*Result, error)
	GetStatus(ctx context.Context, key string) (*Result, error)
	Reset(ctx context.Context, key string) error
	Close() error
}
