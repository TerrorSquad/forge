# Feature 031: Run Single Tool

## Summary
Add a `--tool` flag to `booster run <hook>` so developers can run exactly one
configured tool without executing the full hook chain.

## Motivation
When iterating on a failing static-analysis rule, re-running the entire
pre-commit hook (including formatter, linter, type-checker) wastes 10–30
seconds per attempt. `booster run pre-commit --tool=phpstan` runs only phpstan
against the currently staged files, giving sub-second feedback loops.

## CLI Interface

```
booster run <hook> --tool <name>           # run one tool
booster run <hook> --tool <name>,<name>    # run multiple tools (comma-separated)
booster run <hook> --group <name>          # run all tools in a group
booster run <hook> --skip-tool <name>      # run all except the named tool(s)
```

### Examples

```bash
# Run only phpstan on staged PHP files
booster run pre-commit --tool phpstan

# Run the entire analysis group
booster run pre-commit --group analysis

# Run everything except the slow psalm
booster run pre-commit --skip-tool psalm

# Run two tools in the pre-push hook
booster run pre-push --tool tests,spectral --all-files
```

## Functional Requirements

1. `--tool` accepts a comma-separated list of tool names. Unknown names produce
   a clear error: `unknown tool "foo" in hook pre-commit`.
2. `--group` filters to all tools whose `group` field matches.
3. `--skip-tool` is the inverse of `--tool` — runs all tools except the named
   ones. Useful for "run everything but the slow one".
4. When `--tool` is combined with `--all-files`, file filtering is still
   applied (extensions/patterns for that tool).
5. `depends_on` relationships are respected: if tool B depends on A and only B
   is requested, A is also run automatically (with a note: `  (dependency) A`).
6. Cache is still consulted unless `--no-cache` is passed.

## Error Handling

```
$ booster run pre-commit --tool nonexistent
Error: tool "nonexistent" not found in hook "pre-commit"
Available tools: ecs, rector, phpstan, psalm, deptrac
```

## Out of Scope
- Running tools from different hooks in a single invocation.
- Overriding tool config from CLI flags.
