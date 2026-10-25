---
name: app-performance-testing
description: >
  Investigate why this app's HTTP API endpoints are slow and how to make them faster. Discovers
  the endpoints, measures their response times, works out which layer the time goes to (network,
  middleware, handler code, database, external calls, cache, serialization) using Grafana metrics,
  traces, logs and profiles plus the code, and produces a ranked list of fixes. Use whenever the
  user asks why the app or an endpoint is slow, wants the API to feel snappier, mentions latency,
  p95, slow queries, N+1, or wants a performance check or audit of the backend, even if they
  don't say "performance testing".
---

# App performance testing

Measure single-user endpoint speed, find the layer the time goes to, and rank fixes. The default
is one request at a time after warm-up, not load. Everything is evidence-led: each conclusion is
labelled **measured** (with the evidence) or **inferred from code**, never one passed off as the
other.

Grafana work is delegated to the Grafana skills when installed (`promql`, `loki`, `tempo`, `k6`,
`pyroscope`, `performance-rca`). Load the one for the query language you are about to write
instead of guessing syntax.

## Per-repo files

The skill is generic; each app keeps its own facts under `<repo>/.claude/perf/`:

| File | Holds |
|---|---|
| `app-context.md` | stack, how to start the app, base URLs, env detection, auth, service names, datasource UIDs, production log sources, seeder command, known findings |
| `grafana-queries.md` | ready-to-run `gcx` commands checked against this app's real metric, label and service names |
| `endpoints.json` | default endpoint list for `measure_endpoints.py` |
| `gcx-env.sh` | optional wrapper that feeds `gcx` its credentials from `.env` |

Reports go to `<repo>/.perf-reports/` (gitignored), next to the seeder's `seed-manifest.json`.

**No `app-context.md` yet?** Build it before measuring: read the code for framework, route
definitions, start command, auth and telemetry (OpenTelemetry, Prometheus, Pyroscope,
`service.name`); read `.env` variable **names** only; ask the user what the code can't show
(Grafana Cloud or self-hosted, which signals Grafana receives, how to tell laptop from
production, where production logs live). Run `gcx datasources list` and one real query per
signal, then write `app-context.md` and `grafana-queries.md` from what actually came back.
Seeding (step 2) needs an app-side seeder; if there is none, offer to build one with the guard
rules in *Safety*.

## Workflow

### 1. Load context and detect the environment

Read `.claude/perf/app-context.md`. Detect the mode with the method it records and say which at
the start of the run. If a fact is missing, find it or ask, then update the file.

- **Local mode** (the laptop): the full workflow, including seeding, temporary instrumentation
  (with the user's OK), light concurrency checks and applying fixes.
- **Production mode** (the VPS, or whenever detection is unclear): read-only investigation with
  real data. State the limits once, briefly:
  - Allowed: single-user GET measurements with a small sample (3 warm-up + 5 measured); `gcx`
    reads; the VPS-only log sources in `app-context.md`, read-only and time-boxed (`--since`,
    `tail -n`, `grep -m`); `EXPLAIN` (never `EXPLAIN ANALYZE`) through a read-only client.
  - Not allowed: load or concurrency tests, write requests, changing config, log levels or code,
    restarts, temporary instrumentation, seeding.
  - Fixes are **proposed, not applied**: code edited on the VPS conflicts with the git sync.
    Put them in the report for the laptop to make.
  - Never print secrets or personal data from logs; summarise instead.

### 2. Prepare data (local mode only)

Check `.perf-reports/seed-manifest.json` and current row counts. If the database is empty, small,
or at a different scale from the one the user wants, offer to reseed with the command in
`app-context.md` ("reseed at 10×? wipes dev data"), and run it after one confirmation per run.
Record scale and seed in the report; only compare runs at the same scale and seed. Never seed in
production mode; the seeder refuses too.

### 3. Find endpoints

List routes from the code (method, path, handler file). Start from `.claude/perf/endpoints.json`
and add anything new. Fill path parameters:

- Local: known IDs from the manifest, measuring each relevant endpoint for both the **typical**
  and the **heavy** record. The heavy case is where slowdowns hide.
- Production: IDs found read-only in Grafana or logs, or asked for.

Show the list and let the user narrow it.

### 4. Measure from the client side

```bash
python3 <skill-dir>/scripts/measure_endpoints.py .claude/perf/endpoints.json \
  --base-url <base url from app-context.md> --manifest .perf-reports/seed-manifest.json \
  --warmup 3 --samples 10 -o .perf-reports/<YYYY-MM-DD-HHMM>-raw.json
```

It records DNS, connect, TLS, TTFB, download and total per request, p50/p95/max, size, status and
each endpoint's UTC window (to query Grafana for the same window), and flags TTFB p50 >200 ms
medium, >500 ms high, >1000 ms critical. It refuses non-GET/HEAD/OPTIONS without
`--allow-writes`, which it accepts only for local hosts. In production use
`--warmup 3 --samples 5` and pass a session via `--cookie-env VAR` (the user sets the variable;
never echo it).

### 5. Split the time into layers

For each flagged endpoint, take the best evidence available, in this order:

1. **Client timings.** DNS/connect/TLS is network. A big TTFB-to-total gap with a large body
   points to payload or serialization.
2. **Traces** (`tempo` skill). Find traces for the route in the window; sum child-span time by
   kind (DB, HTTP client, cache, other). The remainder is the handler's own code. Count DB spans
   per request: many near-identical queries means N+1.
3. **Metrics** (`promql` skill): request duration by route, DB time, pool waits, CPU, GC.
4. **Profiles** (`performance-rca`, or `pyroscope`): when handler time dominates, find hot
   functions and map them to files.
5. **Logs** (`loki` skill): per-request access-log durations, slow-query lines, timeouts and
   retries in the window. In production, access logs from real users are often better evidence
   than this skill's handful of requests.
6. **VPS-only logs** (production mode), from `app-context.md`.

Use `-o json` and narrow windows (measurement window plus a small margin). Check flags with
`gcx help-tree <signal> -o text` rather than guessing.

**No telemetry for a layer?** Read the request path in code (route → middleware → service → data
access → external calls → serialization) with `references/bottleneck-checklist.md`, and `EXPLAIN`
the queries. Temporary query logging is allowed locally; revert it and list it in the report.

### 6. Find causes and fixes

Use `references/bottleneck-checklist.md`: N+1, missing indexes, sequential independent calls,
missing caching, over-fetching and big payloads, pool limits, CPU-heavy code, blocking I/O. Each
fix names the file and line, sketches the code, estimates the gain ("21 queries → 1, roughly
−150 ms") and the effort (S/M/L).

### 7. Write the report

`.perf-reports/<YYYY-MM-DD-HHMM>-report.md`, with the raw JSON beside it:

- **Header:** mode; in local mode, data scale and seed.
- **Summary:** the top 3 wins.
- **Endpoint table:** method, path, record (typical/heavy), p50, p95, dominant layer, severity.
  Every measured endpoint has a p95.
- **One section per finding**, ranked by impact against effort: evidence (measured or inferred)
  → cause → fix (file path) → expected gain → effort.
- **Temporary changes** made and reverted, if any.

### 8. Optional checks (local mode, when asked)

- **Before/after:** re-measure after a fix at the same scale and seed.
- **Scaling test:** seed 1×, measure, reseed 10×, measure; rank endpoints by growth. About 10×
  slower at 10× data means work proportional to table size (missing index, no pagination, N+1);
  roughly flat is fine. Often the clearest single signal.
- **Light concurrency:** 1 → 5 → 10 users with k6 (`k6` skill) to spot pool or lock limits.
- **Production-like resources:** run the app under the Compose override named in
  `app-context.md` that caps CPU and memory to the VPS's size.

## Safety

These protect production data and secrets, so they hold even when the user asks casually:

- **Read-only methods by default.** GET, HEAD and OPTIONS need no permission. POST, PUT, PATCH and
  DELETE change data: ask per run, local or dev only, test data only, never production. Asked to
  benchmark writes against production, decline that part and offer local with seeded data.
- **Single-user speed by default.** Load needs an explicit request and a confirmed local target.
  Never put load on production.
- **Seeding is local only.** It wipes data. Check the mode before calling the seeder; the seeder
  itself must refuse unless the environment says local, the DB host is local and an explicit
  wipe flag is passed.
- **Grafana is read-only.** Only `gcx` query, label, list and help commands. Never `gcx
  resources`, `api`, `alert`, `login`, `datasources create/update/delete` or any `adaptive`
  subcommand. Project deny rules and a Viewer token back this up.
- **No secrets in output.** Never print `.env` values or session cookies; mask tokens in reports;
  read credentials from the environment at runtime instead of copying them into files.
- **Undo temporary changes** (logging flags, timing code) before finishing, and list them.
