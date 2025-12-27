package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/AaronBrownDev/distributed-rate-limiter/internal/storage/redis"
	"github.com/AaronBrownDev/distributed-rate-limiter/internal/usecase"
)

func BenchmarkCheckRateLimit_8Goroutines(b *testing.B) {
	benchmarkConcurrent(b, 8)
}

func BenchmarkCheckRateLimit_64Goroutines(b *testing.B) {
	benchmarkConcurrent(b, 64)
}

func BenchmarkCheckRateLimit_512Goroutines(b *testing.B) {
	benchmarkConcurrent(b, 512)
}

func benchmarkConcurrent(b *testing.B, parallelism int) {
	ctx := context.Background()
	storage, _ := redis.NewRedisStorage(ctx, "redis:6379", "bench:")
	defer storage.Close()

	service := usecase.NewRateLimiterService(storage)

	b.SetParallelism(parallelism)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			service.CheckRateLimit(ctx, "bench-key", 1000000, 60*time.Second, 1)
		}
	})
}
