//go:build integration
// +build integration

package redis_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AaronBrownDev/distributed-rate-limiter/internal/storage/redis"
	"github.com/AaronBrownDev/distributed-rate-limiter/internal/usecase"
)

func TestIntegration_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	storage, err := redis.NewRedisStorage(ctx, "redis:6379", "test:")
	if err != nil {
		t.Fatalf("failed to connect to Redis: %v", err)
	}

	// key config
	key := "integration-test"

	t.Cleanup(func() {
		defer storage.Close()

		cleanupCtx := context.Background()
		if err = storage.Reset(cleanupCtx, key); err != nil {
			t.Logf("failed to delete the key %s: %v", key, err)
		}
	})

	service := usecase.NewRateLimiterService(storage)

	// config
	var limit int64 = 100000
	var concurrentRequests int64 = 150000

	var wg sync.WaitGroup
	allowedCount := atomic.Int64{}
	deniedCount := atomic.Int64{}
	errorCount := atomic.Int64{}

	// Loop tests - should be allowed
	for i := int64(1); i <= concurrentRequests; i++ {
		wg.Add(1)
		go func(reqNum int64) {
			defer wg.Done()
			result, err := service.CheckRateLimit(ctx, key, limit, 60*time.Second, 1)
			if err != nil {
				t.Errorf("request %d failed: %v", i, err)
				errorCount.Add(1)
				return
			}

			if result.Allowed {
				allowedCount.Add(1)
			} else {
				deniedCount.Add(1)
			}
		}(i)
	}

	wg.Wait()

	if allowedCount.Load() != limit {
		t.Errorf("Race Condition: expected exactly %d allowed, got %d", limit, allowedCount.Load())
	}

	expectedDenied := concurrentRequests - limit
	if deniedCount.Load() != expectedDenied {
		t.Errorf("expected %d denied, got %d", expectedDenied, deniedCount.Load())
	}

	if errorCount.Load() > 0 {
		t.Errorf("had %d errors during concurrent requests", errorCount.Load())
	}

	t.Logf("Summary: %d allowed, %d denied, %d errors out of %d requests",
		allowedCount.Load(), deniedCount.Load(), errorCount.Load(), concurrentRequests)
}
