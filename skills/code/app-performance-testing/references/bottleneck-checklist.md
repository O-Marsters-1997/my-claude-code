# Bottleneck checklist

> Sources: jeremylongshore/claude-code-plugins-plus-skills `running-performance-tests` (MIT, © 2025 Jeremy Longshore); tonone-ai/tonone `spine-perf` (MIT, © 2026 tonone-ai).

**Walk the request path layer by layer.** For each layer, note what it costs and how you know.
1. Route handler / controller
2. Middleware on this path (auth, logging, validation, rate limiting)
3. Service / business logic
4. Data access (ORM calls, raw SQL)
5. External API calls
6. Response serialization

## Database
- **N+1 queries:** a list is fetched, then each item triggers another query. Look for lazy-loaded relations accessed in loops, or `.map`/`forEach` over ORM objects. Fix with eager loading, a join or a batched query. Estimate the gain: N+1 with 100 items means 101 queries → 1.
- **Missing indexes:** check that columns used in WHERE, JOIN and ORDER BY are indexed, and that multi-column filters have composite indexes. Check the migrations or schema, and run EXPLAIN on suspect queries locally.
- **Slow queries:** check slow-query logs or DB metrics for the measurement window.
- **Over-fetching:** `SELECT *` when only a few fields are needed; no pagination on lists.

## Running work one after another that doesn't depend on each other
- Independent DB queries or external calls made sequentially → run them in parallel.
- File I/O or heavy computation on the request path → move it to a background job or precompute it.
- A new connection opened per request → reuse connections or use a pool.

## Caching
- Data that's read often and rarely changes (config, feature flags, profiles), expensive aggregations, and external responses that can be slightly stale.
- For each one, decide where to cache it (in memory, Redis, or HTTP cache headers / ETag), how long it lives, and how it's invalidated.

## Serialization / payload
- Large nested objects when the client needs only part of them; missing field selection; no compression; a large gap between time to first byte and total time.

## Resource limits (these mostly appear under concurrency, but are worth checking)
- Connection-pool exhaustion (DB or HTTP client): timeouts begin at a particular concurrency level.
- CPU-bound code: compare with profiling data.
- Memory growth (RSS rising over time) and garbage-collection pauses.

## Symptoms → likely cause → fix

| Symptom | Likely cause | Fix |
|---|---|---|
| Connection reset by peer | Server or load balancer dropping connections | Check max-connections settings, pool size and keep-alive |
| Timeouts start at a certain concurrency | Thread pool or DB connection pool exhausted | Check pool usage, size the pool, optimise slow queries |
| Results vary between runs | Cache warm-up, GC pauses, noisy neighbours | Warm up first, run 3 times and compare, use a quiet machine |
| Every request is served from cache | Test inputs are too uniform | Vary parameters and IDs between requests |

## Further reading (not copied: the first is mostly about heavy load, the second has no licence)
- Reading the latency curve under load: https://github.com/OneWave-AI/claude-skills/blob/main/api-load-tester/references/metrics-interpretation.md
- Interpreting results and grading severity: https://github.com/rcampos09/performance-testing-skills/blob/main/skills/performance-report-analysis/SKILL.md
