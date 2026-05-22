# Feature 015: Parallel Tool Execution

## Summary
Run independent tools within a hook concurrently, reducing wall-clock time for
hooks with multiple linters/formatters.

## Motivation
Sequential execution serialises gofmt + govet + golangci-lint even though they
are fully independent. On a large repo this adds several seconds per commit.
Parallel execution reduces total hook time to the slowest single tool.

## Functional Requirements

1. Parallel mode is opt-in at the hook or global level:
   ```toml
   [execution]
   parallel = true          # global default

   [hooks.pre-commit]
   parallel = true          # hook-level override
   ```
2. Tools that declare `depends_on = ["gofmt"]` run only after their
   dependencies complete successfully.
3. If any tool fails, all still-running goroutines are waited for before
   returning the combined error.
4. `on_failure = "stop"` applies globally — if one tool fails, no new tools
   are launched (in-flight tools finish).
5. Output from parallel tools must not interleave — buffer each tool's
   stdout/stderr and flush atomically when the tool exits.
6. Per-tool timing is still reported even in parallel mode.

## Config Schema

```toml
[execution]
parallel = true

[hooks.pre-commit.tools.golangci-lint]
depends_on = ["gofmt"]     # waits for gofmt before starting
```

## Non-Functional Requirements
- No additional external dependencies (use `sync.WaitGroup` / `errgroup`).
- Sequential mode remains the default (backward compatible).
- Parallel mode must be safe with `restage = true` (restage only after all
  parallel tools in a group finish).

## Out of Scope
- Cross-hook parallelism.
- Worker pool size configuration (use GOMAXPROCS).
