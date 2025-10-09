package redis

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/AaronBrownDev/distributed-rate-limiter/internal/storage"
	"github.com/redis/go-redis/v9"
)

// RedisStorage implements storage.RateLimitStorage
type RedisStorage struct {
	client    *redis.Client
	keyPrefix string
}

// TODO: Look into Lua scripts for optimizing Redis operations

// NewRedisStorage
func NewRedisStorage(ctx context.Context, addr, keyPrefix string) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,

		PoolSize:     10,
		MinIdleConns: 5,
	})

	response := client.Ping(ctx)
	if err := response.Err(); err != nil {
		return nil, err
	}

	return &RedisStorage{
		client:    client,
		keyPrefix: keyPrefix,
	}, nil
}

// CheckAndUpdate checks if a request is allowed and updates the counter
func (rs *RedisStorage) CheckAndUpdate(ctx context.Context, key string, limit int64, window time.Duration, cost int64) (*storage.Result, error) {
	// NOTE: This implementation has a small race condition between INCR and EXPIRE.
	// See ADR-0001 for rationale on accepting this trade-off.
	// Future version will use Lua scripts for true atomicity.

	// Build Redis key
	redisKey := rs.formatKey(key)

	// Increments count by cost
	count, err := rs.client.IncrBy(ctx, redisKey, cost).Result()
	if err != nil {
		return nil, err
	}

	var resetAt time.Time
	if count == cost {
		// If first request, set expiration
		if err := rs.client.Expire(ctx, redisKey, window).Err(); err != nil {
			return nil, err
		}
		resetAt = time.Now().Add(window)
	} else {
		// Calculate reset time
		ttl, err := rs.client.TTL(ctx, redisKey).Result()
		if err != nil {
			return nil, err
		}
		resetAt = time.Now().Add(ttl)
	}

	// Check the limit
	allowed := count <= limit

	// Calculate remaining tokens
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	return &storage.Result{
		Allowed:   allowed,
		Remaining: remaining,
		ResetAt:   resetAt,
		Limit:     limit,
	}, nil
}

// GetStatus checks current status without modifying the counter
func (rs *RedisStorage) GetStatus(ctx context.Context, key string, limit int64) (*storage.Result, error) {
	// TODO: Store limit in Redis. Take out of parameter.

	// Build Redis key
	redisKey := rs.formatKey(key)

	// Get current count
	getOutput, err := rs.client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		// Key doesn't exist so just give default result
		// TODO: Look into alternative approaches for determining ResetAt if needed
		return &storage.Result{
			Allowed:   true,
			Remaining: limit,
			ResetAt:   time.Now(),
			Limit:     limit,
		}, nil
	} else if err != nil {
		return nil, err
	}
	count, err := strconv.ParseInt(getOutput, 10, 64)
	if err != nil {
		count = 0 // TODO: Could swap to return nil, err. Should look into that
	}

	// Check the limit
	allowed := count <= limit

	// Calculate remaining tokens
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	// Get time until reset
	ttl, err := rs.client.TTL(ctx, redisKey).Result()
	if err != nil {
		return nil, err
	}
	resetAt := time.Now().Add(ttl)

	return &storage.Result{
		Allowed:   allowed,
		Remaining: remaining,
		ResetAt:   resetAt,
		Limit:     limit,
	}, nil
}

// Reset clears the rate limiter for an identifier
func (rs *RedisStorage) Reset(ctx context.Context, key string) error {

	// Build Redis key
	redisKey := rs.formatKey(key)

	// Delete the key
	keysDeleted, err := rs.client.Del(ctx, redisKey).Result()
	if err != nil {
		return err
	}

	if keysDeleted == 0 {
		return errors.New("no key was deleted")
	}
	return nil
}

// Close cleans up connections when shutting down
func (rs *RedisStorage) Close() error {
	return rs.client.Close()
}

// formatKey consistently formats Redis keys
func (rs *RedisStorage) formatKey(identifier string) string {
	return rs.keyPrefix + identifier
}
