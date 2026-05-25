# Feature 020: Shell Completions

## Summary
Generate shell completion scripts for bash, zsh, and fish via
`forge completion <shell>`.

## Motivation
Tab-completion for subcommands, flags, and hook names reduces typos and makes
forge feel like a first-class CLI tool.

## Functional Requirements

1. `forge completion bash` prints a bash completion script.
2. `forge completion zsh` prints a zsh completion script (compdef style).
3. `forge completion fish` prints a fish completion script.
4. Completions cover:
   - All subcommands (`init`, `install`, `uninstall`, `run`, `migrate`,
     `doctor`, `version`, `completion`)
   - Flags for each subcommand
   - `forge run <TAB>` completes known hook names (`pre-commit`,
     `commit-msg`, `pre-push`, `prepare-commit-msg`, `post-commit`)
   - `forge init --preset <TAB>` completes preset names
   - `forge completion <TAB>` completes shell names
5. Installation instructions are printed as a comment at the top of each
   script.

## Example

```sh
forge completion zsh > ~/.zsh/completions/_forge
forge completion bash > /etc/bash_completion.d/forge
forge completion fish > ~/.config/fish/completions/forge.fish
```

## Non-Functional Requirements
- No external deps — emit static completion scripts as string constants.
- Scripts are generated from a single source-of-truth map so adding a new
  subcommand automatically extends completions.
