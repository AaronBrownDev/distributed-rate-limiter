# ADR-0002: gRPC API Design

## Date
2025-12-12

## Status
Accepted

---

## Context

This ADR applies to the proto implementation currently located at 'api/proto/ratelimiter/v1'. We need to expose our rate limiter through a gRPC API. 

Key design questions:
1. What operations should we expose?
2. What data should responses include?
3. Should we include `current` (int64) in GetStatus?
4. How should we handle errors?

---

## Decision

**Three RPC methods:**
- `CheckRateLimit` - Check if allowed AND consume tokens
- `GetStatus` - Query status WITHOUT consuming tokens  
- `ResetLimit` - Clear rate limit for a key

**Response design:**
- Include `current` in GetStatusResponse
  - `current` enables dashboards to show "80/100 requests used"
  - It can be calculated by `limit` - `remaining`, but saves client from doing math 
  - it makes the API easier to use

- Add `retry_after_seconds` to CheckRateLimitResponse
  - Calculated from `reset_at` in the gRPC server
  - Saves clients from doing time math

**Error handling:**
- Use gRPC status codes instead of error fields in messages
- Example: ResetLimitResponse is empty - errors communicated via status codes

**Temporary design:**
- GetStatusRequest includes `limit` parameter until we store limits in Redis

---

## Consequences

### Positive
- Complete information for different client use cases
- No time math or calculations required by clients
- Follows gRPC best practices
- Easy to add more fields later

### Negative
- Some data redundancy in responses
- Temporary coupling with `limit` parameter in GetStatus
- Slightly larger response payloads

---

## Alternatives Considered

**Minimal responses (only `allowed` and `remaining`):**
- Rejected: Dashboards couldn't show usage progress without extra math

**Error fields in messages:**
- Rejected: Not idiomatic gRPC, breaks interceptors and error handling patterns

**Just `current` without `allowed`:**
- Rejected: Clients would need to calculate `current < limit` themselves