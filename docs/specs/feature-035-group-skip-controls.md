# Feature 035: Group-Level Skip Controls

## Summary
Allow skipping entire tool groups via `SKIP_GROUP_<NAME>=1` environment
variables, complementing the existing per-tool `SKIP_<TOOL>=1` controls.

## Motivation
With large hook configs, suppressing a logical category of checks (e.g. all
static analysis tools) currently requires setting multiple `SKIP_*` vars:

```bash
SKIP_PHPSTAN=1 SKIP_PSALM=1 SKIP_DEPTRAC=1 git commit -m "wip"
```

When a `group = "analysis"` is already declared on each tool, one variable
should be enough:

```bash
SKIP_GROUP_ANALYSIS=1 git commit -m "wip"
```

## Functional Requirements

1. Group names in env vars are uppercased: `group = "analysis"` →
   `SKIP_GROUP_ANALYSIS=1`.
2. Group skip is additive: `SKIP_GROUP_ANALYSIS=1 SKIP_ECS=1` skips the
   analysis group AND ecs (even if ecs is in a different group).
3. When a tool is skipped by group, its row shows `~` (same as per-tool skip)
   with a note: `(group: analysis)`.
4. `SKIP_GROUP_*` is documented by `forge doctor` and listed in the hook
   output when `FORGE_DEBUG=1`.
5. Group names are case-insensitive in env var matching (`SKIP_GROUP_Analysis`
   and `SKIP_GROUP_ANALYSIS` are equivalent).

## Visual Design

```
$ SKIP_GROUP_ANALYSIS=1 git commit -m "wip: prototype"

pre-commit
  ✓  ecs                   340ms
  ~  phpstan                     (group: analysis)
  ~  psalm                       (group: analysis)
  ~  deptrac                     (group: analysis)
  ✓  multiline-attributes  120ms

3 passed · 2 skipped · total 465ms
```

## Error Handling

Unknown group names produce a warning (not an error):
```
⚠ SKIP_GROUP_UNKNOWN=1: no tools belong to group "unknown" in pre-commit
```

## `forge doctor` Integration

`forge doctor` lists available group skip vars for the current config:

```
Available SKIP_GROUP_* variables (pre-commit):
  SKIP_GROUP_FORMAT    skips: ecs, rector, multiline-attributes
  SKIP_GROUP_ANALYSIS  skips: phpstan, psalm, deptrac
  SKIP_GROUP_ARTIFACTS skips: deptrac-image
```

## Out of Scope
- `ONLY_GROUP_*` (run only a specific group) — use `forge run --group` from
  feature-031 for that use case.
- Nested groups or group hierarchies.
