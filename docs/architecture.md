# Architecture

How a git operation flows through forge, and what each package owns.

## Request pipeline

```mermaid
flowchart TD
    git["git commit / push / merge"] --> shim[".forge/hooks/&lt;hook&gt; shim"]
    shim --> cli["forge run &lt;hook&gt;<br/>(cmd/forge → internal/forge/app.go)"]
    cli --> run["runner.RunHookWithOptions"]

    run --> root["git.DetectRepoRoot"]
    run --> envf["load .git-hooks.env / .env"]
    run --> skip{"SKIP_* env<br/>or merge/rebase<br/>in progress?"}
    skip -->|yes| stop["skip hook"]
    skip -->|no| load["config.LoadConfig<br/>(repo forge.toml + global merge)"]

    load --> policy{"commit-msg /<br/>prepare-commit-msg?"}
    policy -->|yes| pol["apply commit-message policy<br/>(conventional, ticket footer)"]
    policy -->|no| ws{"workspace<br/>members match<br/>staged files?"}
    ws -->|yes| member["run hook per member<br/>(member forge.toml)"]
    ws -->|no| files["gather files<br/>(staged, or all-tracked for --all-files)"]

    files --> pre{"all enabled tool<br/>binaries present?"}
    pre -->|no| prefail["fail hook<br/>(install / SKIP_* / disable)"]
    pre -->|yes| mode{"parallel mode?"}
    mode -->|no| seq["runHookCfg<br/>(sequential)"]
    mode -->|yes| par["runHookCfgParallel<br/>(waves by depends_on)"]

    seq --> tool
    par --> tool

    subgraph tool["per tool"]
        filter["filter files<br/>(extensions, glob patterns)"] --> backend["backend.ResolveBackend<br/>host / ddev / docker"]
        backend --> cache{"cache hit?"}
        cache -->|yes| skiptool["skip (cached)"]
        cache -->|no| exec["execute<br/>(timeout, env, check-mode)"]
        exec --> after["on success:<br/>restage / stage_outputs / update cache"]
    end
```

## Package map

| Package | Responsibility |
|---------|----------------|
| `cmd/forge` | Binary entry point; calls `forge.Run(args)` |
| `internal/forge` (`app.go`) | CLI dispatch for every subcommand; also `install`, `uninstall`, `doctor`, `validate`, `completion` |
| `internal/forge/config` | Load/parse/merge `forge.toml`, presets, Husky migration, remote presets, global-config merge |
| `internal/forge/runner` | Core execution: file filtering, sequential + parallel runners, commit-message policy, run cache, workspace routing |
| `internal/forge/backend` | Execution abstraction — `HostBackend`, `DdevBackend`, `DockerBackend`; DDEV auto-detection |
| `internal/forge/git` | Repo-root detection, staged/tracked file listing, `git add`, sequencer-state checks |
| `internal/forge/ui` | Terminal output (colors, spinners) and the plain CI variant |
| `internal/forge/update` | Self-update from GitHub releases |

## Key invariants

- **Tool order** follows declaration order in `forge.toml`. Go maps don't preserve order, so `config.parseHookToolOrder` regex-scans the raw TOML and feeds `OrderedToolNames()` — never rely on map iteration for ordering.
- **Backend resolution precedence**: per-tool `backend` → `[execution].default_backend` → DDEV auto-detect (if the container is running) → host.
- **Preflight before execution**: every enabled (non-skipped) tool's binary must resolve before any tool runs; a missing host binary aborts the whole hook. Container backends are assumed to provide their own binaries.
- **Mutations happen only on success**: `restage` and `stage_outputs` run after a tool passes (never in `--check` mode); the run cache is only written for passing tools.
- **Config precedence**: `FORGE_CONFIG` → repo `forge.toml`, with the global user config (`~/.config/forge/config.toml`) merged underneath — repo values always win.
