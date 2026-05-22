# Feature 020: Shell Completions

## Summary
Generate shell completion scripts for bash, zsh, and fish via
`booster completion <shell>`.

## Motivation
Tab-completion for subcommands, flags, and hook names reduces typos and makes
booster feel like a first-class CLI tool.

## Functional Requirements

1. `booster completion bash` prints a bash completion script.
2. `booster completion zsh` prints a zsh completion script (compdef style).
3. `booster completion fish` prints a fish completion script.
4. Completions cover:
   - All subcommands (`init`, `install`, `uninstall`, `run`, `migrate`,
     `doctor`, `version`, `completion`)
   - Flags for each subcommand
   - `booster run <TAB>` completes known hook names (`pre-commit`,
     `commit-msg`, `pre-push`, `prepare-commit-msg`, `post-commit`)
   - `booster init --preset <TAB>` completes preset names
   - `booster completion <TAB>` completes shell names
5. Installation instructions are printed as a comment at the top of each
   script.

## Example

```sh
booster completion zsh > ~/.zsh/completions/_booster
booster completion bash > /etc/bash_completion.d/booster
booster completion fish > ~/.config/fish/completions/booster.fish
```

## Non-Functional Requirements
- No external deps — emit static completion scripts as string constants.
- Scripts are generated from a single source-of-truth map so adding a new
  subcommand automatically extends completions.
