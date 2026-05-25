package forge

import (
	"fmt"
	"strings"
)

// knownHooks lists all hook names booster supports.
var knownHooks = []string{
	"pre-commit",
	"commit-msg",
	"pre-push",
	"prepare-commit-msg",
	"post-commit",
}

// subcommands lists all top-level booster subcommands.
var subcommands = []string{
	"init",
	"install",
	"uninstall",
	"run",
	"migrate",
	"doctor",
	"version",
	"completion",
	"help",
}

// GenerateCompletion returns a shell completion script for the given shell.
// Supported values: "bash", "zsh", "fish".
func GenerateCompletion(shell string) (string, error) {
	switch strings.ToLower(shell) {
	case "bash":
		return bashCompletion(), nil
	case "zsh":
		return zshCompletion(), nil
	case "fish":
		return fishCompletion(), nil
	default:
		return "", fmt.Errorf("unsupported shell %q; supported: bash, zsh, fish", shell)
	}
}

func bashCompletion() string {
	hooks := strings.Join(knownHooks, " ")
	presets := strings.Join(ListPresets(), " ")
	cmds := strings.Join(subcommands, " ")

	return fmt.Sprintf(`# booster bash completion
# Install: source this file or add to /etc/bash_completion.d/booster
# Quick install: booster completion bash > /etc/bash_completion.d/booster

_%[1]s_completions() {
  local cur prev words
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"

  local subcommands="%[2]s"
  local hooks="%[3]s"
  local presets="%[4]s"
  local shells="bash zsh fish"

  if [[ ${COMP_CWORD} -eq 1 ]]; then
    COMPREPLY=( $(compgen -W "${subcommands}" -- "${cur}") )
    return 0
  fi

  case "${COMP_WORDS[1]}" in
    run)
      if [[ ${COMP_CWORD} -eq 2 ]]; then
        COMPREPLY=( $(compgen -W "${hooks}" -- "${cur}") )
      else
        COMPREPLY=( $(compgen -W "--all-files --edit" -- "${cur}") )
      fi
      ;;
    init)
      COMPREPLY=( $(compgen -W "--preset --force --list-presets" -- "${cur}") )
      if [[ "${prev}" == "--preset" ]]; then
        COMPREPLY=( $(compgen -W "${presets}" -- "${cur}") )
      fi
      ;;
    migrate)
      COMPREPLY=( $(compgen -W "--from --to" -- "${cur}") )
      ;;
    completion)
      COMPREPLY=( $(compgen -W "${shells}" -- "${cur}") )
      ;;
  esac
}

complete -F _%[1]s_completions %[1]s
`, "forge", cmds, hooks, presets)
}

func zshCompletion() string {
	hooks := `"` + strings.Join(knownHooks, `" "`) + `"`
	presets := `"` + strings.Join(ListPresets(), `" "`) + `"`

	return fmt.Sprintf(`#compdef booster
# booster zsh completion
# Install: booster completion zsh > "${fpath[1]}/_booster"
# Then run: compinit

_booster() {
  local state

  _arguments \
    '1: :->subcommand' \
    '*: :->args'

  case $state in
    subcommand)
      local subcommands
      subcommands=(%[1]s)
      _describe 'subcommand' subcommands
      ;;
    args)
      case ${words[2]} in
        run)
          if [[ ${#words[@]} -eq 3 ]]; then
            local hooks
            hooks=(%[2]s)
            _describe 'hook' hooks
          else
            _arguments '--all-files[run against all tracked files]' '--edit[commit message file]:file:_files'
          fi
          ;;
        init)
          _arguments \
            '--force[overwrite existing config]' \
            '--list-presets[list available presets]' \
            "--preset[starter preset]:preset:(%[3]s)"
          ;;
        migrate)
          _arguments \
            '--from[source config file]:file:_files' \
            '--to[output path]:file:_files'
          ;;
        completion)
          local shells
          shells=("bash" "zsh" "fish")
          _describe 'shell' shells
          ;;
      esac
      ;;
  esac
}

_booster "$@"
`, `"init" "install" "uninstall" "run" "migrate" "doctor" "version" "completion" "help"`, hooks, presets)
}

func fishCompletion() string {
	var sb strings.Builder

	sb.WriteString("# booster fish completion\n")
	sb.WriteString("# Install: booster completion fish > ~/.config/fish/completions/booster.fish\n\n")

	// Subcommands
	for _, cmd := range subcommands {
		sb.WriteString(fmt.Sprintf("complete -c booster -f -n '__fish_use_subcommand' -a %s\n", cmd))
	}
	sb.WriteString("\n")

	// run <hook>
	for _, hook := range knownHooks {
		sb.WriteString(fmt.Sprintf(
			"complete -c booster -f -n '__fish_seen_subcommand_from run' -a %s -d 'hook'\n", hook))
	}
	sb.WriteString("complete -c booster -n '__fish_seen_subcommand_from run' -l all-files -d 'run against all tracked files'\n")
	sb.WriteString("complete -c booster -n '__fish_seen_subcommand_from run' -l edit -d 'commit message file' -r\n\n")

	// init flags
	sb.WriteString("complete -c booster -n '__fish_seen_subcommand_from init' -l force -d 'overwrite existing config'\n")
	sb.WriteString("complete -c booster -n '__fish_seen_subcommand_from init' -l list-presets -d 'list presets'\n")
	for _, p := range ListPresets() {
		sb.WriteString(fmt.Sprintf(
			"complete -c booster -n '__fish_seen_subcommand_from init' -l preset -a %s -d 'preset'\n", p))
	}
	sb.WriteString("\n")

	// migrate flags
	sb.WriteString("complete -c booster -n '__fish_seen_subcommand_from migrate' -l from -d 'source file' -r\n")
	sb.WriteString("complete -c booster -n '__fish_seen_subcommand_from migrate' -l to -d 'output file' -r\n\n")

	// completion shells
	for _, s := range []string{"bash", "zsh", "fish"} {
		sb.WriteString(fmt.Sprintf(
			"complete -c booster -f -n '__fish_seen_subcommand_from completion' -a %s -d 'shell'\n", s))
	}

	return sb.String()
}
