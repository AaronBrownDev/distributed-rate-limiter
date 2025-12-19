package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/AaronBrownDev/distributed-rate-limiter/internal/storage"
)

type RateLimiterService struct {
	storage storage.RateLimitStorage
}

func NewRateLimiterService(storage storage.RateLimitStorage) *RateLimiterService {
	return &RateLimiterService{
		storage: storage,
	}
}

var (
	// ErrInvalidKey will be returned is key is empty
	ErrInvalidKey = errors.New("input key is invalid")
	// ErrInvalidLimit will be returned if limit <= 0
	ErrInvalidLimit = errors.New("input limit is invalid")
	// ErrInvalidWindow will be returned if window <= 0
	ErrInvalidWindow = errors.New("input window is invalid")
	// ErrInvalidCost will be returned if cost <= 0
	ErrInvalidCost = errors.New("input cost is invalid")
)

// CheckRateLimit validates input and checks if a request is allowed and updates the counter
func (rls *RateLimiterService) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration, cost int64) (*storage.Result, error) {

	// Validate input
	if len(strings.TrimSpace(key)) == 0 {
		return nil, ErrInvalidKey
	}
	if limit <= 0 {
		return nil, ErrInvalidLimit
	}
	if window <= 0 {
		return nil, ErrInvalidWindow
	}
	if cost <= 0 {
		return nil, ErrInvalidCost
	}

	// Call storage layer to check and update rate limit
	return rls.storage.CheckAndUpdate(ctx, key, limit, window, cost)
}

// GetStatus validates input and checks current status without modifying the counter
func (rls *RateLimiterService) GetStatus(ctx context.Context, key string, limit int64) (*storage.Result, error) {

	// Validate input
	if len(strings.TrimSpace(key)) == 0 {
		return nil, ErrInvalidKey
	}
	if limit <= 0 {
		return nil, ErrInvalidLimit
	}

	// TODO: limit parameter might be removed in the future
	// Call storage layer to get status
	return rls.storage.GetStatus(ctx, key, limit)
}

// ResetLimit validates input and clears the rate limiter for the given key
func (rls *RateLimiterService) ResetLimit(ctx context.Context, key string) error {

	// Validate input
	if len(strings.TrimSpace(key)) == 0 {
		return ErrInvalidKey
	}

	// Call storage layer to reset the limit
	return rls.storage.Reset(ctx, key)
}
