# Feature 034: Hook Timing Metrics

## Summary
Track per-tool execution duration across hook runs and surface slow tools,
regressions, and cumulative overhead — so teams can make informed decisions
about what to parallelize, cache, or move to CI.

## Motivation
Over time, pre-commit hooks creep from 2s to 30s as teams add more tools.
Nobody notices the individual 3-second increases. Timing metrics make the
degradation visible before it becomes painful, and help teams decide whether
to enable caching or parallel execution.

## Data Model

Metrics are stored in `.forge/metrics.jsonl` (newline-delimited JSON, one
record per hook run, gitignored by `forge install`).

```json
{
  "ts": "2025-05-22T09:14:33Z",
  "hook": "pre-commit",
  "branch": "feature/my-branch",
  "total_ms": 9340,
  "tools": [
    { "name": "ecs",     "status": "pass",   "duration_ms": 340  },
    { "name": "phpstan", "status": "cached", "duration_ms": 12   },
    { "name": "psalm",   "status": "pass",   "duration_ms": 4200 }
  ]
}
```

## CLI Interface

### `forge metrics`

```
$ forge metrics

pre-commit  (last 30 runs, avg 8.4s)
  ecs           avg  340ms  p95  450ms
  phpstan       avg  3.2s   p95  5.1s    ⚠ slowest tool
  psalm         avg  4.0s   p95  6.3s    ⚠ slowest tool
  deptrac       avg  900ms  p95  1.2s

  💡 Enable cache for phpstan and psalm to save ~7s per commit.
  💡 Enable parallel execution to cut total time to ~6.3s.

pre-push  (last 12 runs, avg 14.2s)
  tests         avg  8.1s   p95  12.0s
  spectral      avg  900ms  p95  1.4s
  openapi-gen   avg  3.4s   p95  4.8s
```

### `forge metrics --tool phpstan`

Show sparkline history for a single tool (last 20 runs).

```
phpstan  (pre-commit)
  3.1  3.2  3.0  3.4  3.1  ⚡  0.0  0.0  3.2  3.1  3.3  3.5  3.8 ⬆
  ↑ cache hits (near zero)    ↑ regression: +0.8s over last 5 runs
```

### `forge metrics reset`

Clear `.forge/metrics.jsonl`.

## Functional Requirements

1. Metrics are appended at the end of every hook run (pass or fail).
2. Append is atomic (write temp + rename) and never blocks or fails the hook.
3. `forge metrics` reads up to the last 100 run records per hook.
4. Slow tool threshold: warn if a tool's p95 exceeds **3s** (configurable).
5. Regression detection: warn if the last 5-run average is >20% higher than
   the previous 5-run average for any tool.
6. Suggestions are generated automatically:
   - Tool avg > 3s and `cache = false` → suggest enabling cache.
   - Total time > 10s and `parallel = false` → suggest parallel execution.
   - A tool's p95 > 2× avg → suggest investigating flakiness.
7. `FORGE_NO_METRICS=1` disables all metric collection.

## Config

```toml
[metrics]
enabled         = true
slow_threshold  = "3s"
max_records     = 200
```

## Out of Scope
- Shipping metrics to a remote service or time-series DB.
- Per-file timing (would require tool cooperation).
- Metrics across different machines / developers.
