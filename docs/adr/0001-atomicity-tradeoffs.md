# ADR-0001: Redis Atomicity vs Simplicity Trade-off
## Date
2025-10-04  

## Status
Accepted  

---

## Context
This ADR applies to the Redis-backed rate limiter implementation currently located at `internal/limiter/storage/redis.go`.
When implementing the rate limiter's `CheckAndUpdate` function, I need to:

1. Increment a counter (`INCR`)  
2. Check if this is the first request  
3. Conditionally set expiration (`EXPIRE`) only for the first request  

Redis pipelines can batch operations but cannot perform conditional logic based on results within the same pipeline.

#### Initial Draft

```go
// CheckAndUpdate checks if a request is allowed and updates the counter
func (rs *RedisStorage) CheckAndUpdate(ctx context.Context, key string, limit int64, window time.Duration, cost int64) (*Result, error) {
    redisKey := rs.formatKey(key)

    // Start pipeline to make sure Redis operations are atomic
    pipeline := rs.client.Pipeline()

    incrByCmd := pipeline.IncrBy(ctx, redisKey, cost)
    count := incrByCmd.Val()

    var resetAt time.Time
    if count == cost { // Doesn't work: pipeline not executed yet
        pipeline.Expire(ctx, redisKey, window)
        resetAt = time.Now().Add(window)
    } else {
        ttlCmd := pipeline.TTL(ctx, redisKey)
        resetAt = time.Now().Add(ttlCmd.Val())
    }

    if _, err := pipeline.Exec(ctx); err != nil {
        return nil, err
    }

    allowed := count <= limit
    remaining := limit - count
    if remaining < 0 {
        remaining = 0
    }

    return &Result{
        Allowed:   allowed,
        Remaining: remaining,
        ResetAt:   resetAt,
        Limit:     limit,
    }, nil
}
```

---

## Decision
I chose to accept a **non-atomic implementation** that separates `INCR` and conditional `EXPIRE`, with the plan to later move to **Lua scripts** for atomicity.

```go
count, _ := rs.client.IncrBy(ctx, key, cost).Result()
if count == cost { // First request
    rs.client.Expire(ctx, key, window) // Separate operation
}
```

---

## Consequences

### Positive
- Simpler code  
- Easy to test and debug  
- Avoids Lua scripting complexity at this stage  

### Negative
- Race condition: `EXPIRE` could fail after successful `INCR`  
- Potential memory leak if keys never expire  
- Not fully atomic  

---

## Alternatives Considered
- **Always EXPIRE**  
  Would reset TTL on every request, turning fixed window into sliding window.  

- **Lua Scripts**  
  Provides true atomicity but adds upfront complexity.  

- **Redis Transactions**  
  Still cannot do conditional logic atomically.  
