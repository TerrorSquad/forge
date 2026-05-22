# Feature 009: File Filtering and Staged-File Pipeline

## Summary
Define the exact pipeline by which booster discovers, filters, and passes files
to each tool during the pre-commit hook.

## Motivation
Running expensive static-analysis tools on the entire repository on every commit
is unacceptably slow. Only staged files that match a tool's declared scope should
be processed.

## Pipeline Steps

```
git diff --cached --name-only --diff-filter=ACMR
        │
        ▼
 extension filter (tool.extensions)
        │
        ▼
 include_patterns filter (glob, optional)
        │
        ▼
 exclude_patterns filter (glob, optional)
        │
        ▼
 [ if empty AND pass_files=true → skip tool ]
        │
        ▼
 run_per_file=false  → run once, all files as trailing args
 run_per_file=true   → run once per file (max 10 concurrent)
        │
        ▼
 pass_files=false    → run once, no file args
```

## Functional Requirements
1. `extensions` match against `filepath.Ext(file)` case-insensitively.
2. `include_patterns` are glob patterns; if empty, all files pass.
3. `exclude_patterns` are glob patterns; matching files are dropped.
4. When `run_per_file = true`, run a subprocess per file, bounded to 10
   concurrent processes.
5. File paths passed to tools are always repo-root-relative slash-separated paths.
6. After a tool with `restage = true` completes, run `git add -- <files>` for
   exactly the files that were passed to the tool.

## Edge Cases
- Empty staged-file list → print info message, skip all tools, exit 0.
- Tool with `pass_files = false` and empty file list → still runs if its
  extension list would have matched (e.g. whole-project analysers).
