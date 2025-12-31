# Performance Benchmarks

## Test Configuration

All benchmarks run with:
```bash
# Statistical validation (10 runs)
go test -tags=integration -bench=. -benchtime=3s -count=10 ./internal/storage/redis/

# Memory profiling
go test -tags=integration -bench=. -benchtime=3s -benchmem ./internal/storage/redis/
```

**Environment:**
- Go: 1.25.0
- Redis: 7.x (Docker container, localhost)
- OS: Linux (dev container)
- Redis connection pool: 10 max connections, 5 min idle
- Test date: December 27, 2025

## Summary

| System | CPU | Threads | Avg Latency | Memory | Allocations |
|--------|-----|---------|-------------|--------|-------------|
| Desktop | Ryzen 5 7600X | 12 | ~9.8 μs | 465 B | 14 |
| Laptop | Ryzen 7 5800H | 16 | ~15.6 μs | 466 B | 14 |

Latency stays consistent across all concurrency levels tested (96 to 8,192 goroutines).

---

## Desktop: AMD Ryzen 5 7600X (6-Core, 12 Threads)

### Statistical Validation (10 runs, 3 seconds each)

```
goos: linux
goarch: amd64
pkg: github.com/AaronBrownDev/distributed-rate-limiter/internal/storage/redis
cpu: AMD Ryzen 5 7600X 6-Core Processor             
BenchmarkCheckRateLimit_8Goroutines-12            354583              9768 ns/op
BenchmarkCheckRateLimit_8Goroutines-12            359977              9854 ns/op
BenchmarkCheckRateLimit_8Goroutines-12            353959              9932 ns/op
BenchmarkCheckRateLimit_8Goroutines-12            373140              9729 ns/op
BenchmarkCheckRateLimit_8Goroutines-12            361118              9817 ns/op
BenchmarkCheckRateLimit_8Goroutines-12            366380              9810 ns/op
BenchmarkCheckRateLimit_8Goroutines-12            357531              9757 ns/op
BenchmarkCheckRateLimit_8Goroutines-12            380426              9823 ns/op
BenchmarkCheckRateLimit_8Goroutines-12            372589              9793 ns/op
BenchmarkCheckRateLimit_8Goroutines-12            369673              9809 ns/op
BenchmarkCheckRateLimit_64Goroutines-12           375380              9740 ns/op
BenchmarkCheckRateLimit_64Goroutines-12           360494              9851 ns/op
BenchmarkCheckRateLimit_64Goroutines-12           373443              9706 ns/op
BenchmarkCheckRateLimit_64Goroutines-12           352621              9757 ns/op
BenchmarkCheckRateLimit_64Goroutines-12           368157              9743 ns/op
BenchmarkCheckRateLimit_64Goroutines-12           367994              9731 ns/op
BenchmarkCheckRateLimit_64Goroutines-12           352634              9740 ns/op
BenchmarkCheckRateLimit_64Goroutines-12           365632              9819 ns/op
BenchmarkCheckRateLimit_64Goroutines-12           365451              9647 ns/op
BenchmarkCheckRateLimit_64Goroutines-12           364033              9697 ns/op
BenchmarkCheckRateLimit_512Goroutines-12          364021              9955 ns/op
BenchmarkCheckRateLimit_512Goroutines-12          356032              9829 ns/op
BenchmarkCheckRateLimit_512Goroutines-12          352065              9821 ns/op
BenchmarkCheckRateLimit_512Goroutines-12          359461              9783 ns/op
BenchmarkCheckRateLimit_512Goroutines-12          351921              9865 ns/op
BenchmarkCheckRateLimit_512Goroutines-12          366271              9856 ns/op
BenchmarkCheckRateLimit_512Goroutines-12          359498              9782 ns/op
BenchmarkCheckRateLimit_512Goroutines-12          359518              9755 ns/op
BenchmarkCheckRateLimit_512Goroutines-12          352933              9833 ns/op
BenchmarkCheckRateLimit_512Goroutines-12          369940              9843 ns/op
PASS
ok      github.com/AaronBrownDev/distributed-rate-limiter/internal/storage/redis        127.475s
```

### Memory Profile

```
BenchmarkCheckRateLimit_8Goroutines-12            373916             10000 ns/op             464 B/op         14 allocs/op
BenchmarkCheckRateLimit_64Goroutines-12           373574              9696 ns/op             465 B/op         14 allocs/op
BenchmarkCheckRateLimit_512Goroutines-12          349051              9846 ns/op             472 B/op         14 allocs/op
```

### Results

**Latency across concurrency levels:**
- 96 goroutines: 9,729 - 9,932 ns (avg: 9,809 ns, variance: 2.1%)
- 768 goroutines: 9,647 - 9,851 ns (avg: 9,743 ns, variance: 2.1%)
- 6,144 goroutines: 9,755 - 9,955 ns (avg: 9,832 ns, variance: 2.0%)

**Observations:**
- Latency stays around ~9.8 μs regardless of concurrency level
- Low variance (<2.5%) indicates consistent performance
- Memory usage: 464-472 bytes per operation, 14 allocations

---

## Laptop: AMD Ryzen 7 5800H (8-Core, 16 Threads)

### Statistical Validation (10 runs, 3 seconds each)

```
goos: linux
goarch: amd64
pkg: github.com/AaronBrownDev/distributed-rate-limiter/internal/storage/redis
cpu: AMD Ryzen 7 5800H with Radeon Graphics         
BenchmarkCheckRateLimit_8Goroutines-16            216296             15423 ns/op
BenchmarkCheckRateLimit_8Goroutines-16            230144             15412 ns/op
BenchmarkCheckRateLimit_8Goroutines-16            235682             15326 ns/op
BenchmarkCheckRateLimit_8Goroutines-16            228748             15539 ns/op
BenchmarkCheckRateLimit_8Goroutines-16            228214             15619 ns/op
BenchmarkCheckRateLimit_8Goroutines-16            227787             15380 ns/op
BenchmarkCheckRateLimit_8Goroutines-16            236509             15560 ns/op
BenchmarkCheckRateLimit_8Goroutines-16            227931             15904 ns/op
BenchmarkCheckRateLimit_8Goroutines-16            224100             15734 ns/op
BenchmarkCheckRateLimit_8Goroutines-16            220687             16024 ns/op
BenchmarkCheckRateLimit_64Goroutines-16           219045             15388 ns/op
BenchmarkCheckRateLimit_64Goroutines-16           236061             15450 ns/op
BenchmarkCheckRateLimit_64Goroutines-16           223297             15407 ns/op
BenchmarkCheckRateLimit_64Goroutines-16           228326             15292 ns/op
BenchmarkCheckRateLimit_64Goroutines-16           224090             15516 ns/op
BenchmarkCheckRateLimit_64Goroutines-16           235989             15428 ns/op
BenchmarkCheckRateLimit_64Goroutines-16           220842             15423 ns/op
BenchmarkCheckRateLimit_64Goroutines-16           236454             15244 ns/op
BenchmarkCheckRateLimit_64Goroutines-16           233566             15429 ns/op
BenchmarkCheckRateLimit_64Goroutines-16           228795             15506 ns/op
BenchmarkCheckRateLimit_512Goroutines-16          223234             15697 ns/op
BenchmarkCheckRateLimit_512Goroutines-16          229507             15658 ns/op
BenchmarkCheckRateLimit_512Goroutines-16          220627             15505 ns/op
BenchmarkCheckRateLimit_512Goroutines-16          228313             15687 ns/op
BenchmarkCheckRateLimit_512Goroutines-16          233217             15747 ns/op
BenchmarkCheckRateLimit_512Goroutines-16          232524             15527 ns/op
BenchmarkCheckRateLimit_512Goroutines-16          228810             15630 ns/op
BenchmarkCheckRateLimit_512Goroutines-16          204517             15805 ns/op
BenchmarkCheckRateLimit_512Goroutines-16          212552             15768 ns/op
BenchmarkCheckRateLimit_512Goroutines-16          216669             15505 ns/op
PASS
ok      github.com/AaronBrownDev/distributed-rate-limiter/internal/storage/redis        137.929s
```

### Memory Profile

```
BenchmarkCheckRateLimit_8Goroutines-16            221916             15663 ns/op             465 B/op         14 allocs/op
BenchmarkCheckRateLimit_64Goroutines-16           230995             15380 ns/op             466 B/op         14 allocs/op
BenchmarkCheckRateLimit_512Goroutines-16          228658             15405 ns/op             481 B/op         14 allocs/op
```

### Results

**Latency across concurrency levels:**
- 128 goroutines: 15,326 - 16,024 ns (avg: 15,592 ns, variance: 4.5%)
- 1,024 goroutines: 15,244 - 15,516 ns (avg: 15,408 ns, variance: 1.8%)
- 8,192 goroutines: 15,505 - 15,805 ns (avg: 15,653 ns, variance: 1.9%)

**Observations:**
- Latency stays around ~15.6 μs regardless of concurrency level
- Variance under 5% indicates stable performance
- Memory usage: 465-481 bytes per operation, 14 allocations

---

## System Comparison

| Metric | Desktop (7600X) | Laptop (5800H) | Difference |
|--------|-----------------|----------------|------------|
| Average Latency | 9.8 μs | 15.6 μs | 37% faster on desktop |
| Variance | 2.0% | 3.1% | Similar |
| Memory Usage | 465 B | 466 B | Identical |
| Allocations | 14 | 14 | Identical |

The desktop's faster latency is due to newer CPU architecture (Zen 4 vs Zen 3) and higher clock speeds. Both systems show the same scaling behavior.

---

## What This Means

**Latency:**
- Desktop: ~10 microseconds per request
- Laptop: ~16 microseconds per request
- For context: database queries typically take 1-10 milliseconds (100-1000x slower)

**Concurrency:**
- Tested from 96 to 8,192 simultaneous goroutines
- Latency stayed flat across the entire range
- No lock contention or blocking detected

**Memory:**
- Each operation allocates ~465 bytes
- 14 memory allocations per request
- Consistent across all concurrency levels

---

## Running the Benchmarks

**Prerequisites:**
- Dev container running (includes Go 1.25 and Redis 7.x)

**Run benchmarks:**

Quick test:
```bash
go test -tags=integration -bench=. -benchtime=3s -benchmem ./internal/storage/redis/
```

Statistical validation (10 runs):
```bash
go test -tags=integration -bench=. -benchtime=3s -count=10 ./internal/storage/redis/
```

**Expected:**
- Latency: 9-16 μs depending on your CPU
- Variance: <5% across runs
- Memory: ~465 bytes per operation
- Allocations: 14 per operation

---

## Understanding the Numbers

**ns/op (nanoseconds per operation):**
- Time for one `CheckRateLimit` call
- 10,000 ns = 10 microseconds = 0.01 milliseconds
- Lower is better

**B/op (bytes per operation):**
- Memory allocated per call
- Includes Redis client overhead and response parsing

**allocs/op (allocations per operation):**
- Number of memory allocations per call
- Each allocation has a small performance cost

**Why flat latency matters:**
- When latency doesn't increase as you add goroutines, it means no lock contention
- This implementation can handle high concurrency without bottlenecks

---

## Summary

- Desktop latency: ~10 μs
- Laptop latency: ~16 μs  
- Memory: ~465 bytes per operation
- Allocations: 14 per operation
- Concurrency: No degradation from 96 to 8,192 goroutines
- Variance: <5% across 10 runs on both systems

The implementation handles high concurrency without lock contention. Performance differences between systems are due to CPU architecture, not the code.