package config

import "sort"

const defaultConfig = `[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.prettier]
command = "prettier"
args = ["--write", "--ignore-unknown"]
type = "node"
extensions = [".js", ".ts", ".json", ".md", ".yml", ".yaml"]
restage = true
group = "format"

[hooks.pre-commit.tools.eslint]
command = "eslint"
args = ["--fix", "--cache", "--no-warn-ignored"]
type = "node"
extensions = [".js", ".jsx", ".ts", ".tsx", ".vue"]
restage = true
group = "lint"

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
append_ticket_footer = true
require_ticket = false

[hooks.pre-push]
enabled = false
`

var presets = map[string]string{
	"node": defaultConfig,
	"php": `[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.ecs]
command = "vendor/bin/ecs"
args = ["check", "--fix"]
type = "php"
extensions = [".php"]
restage = true
group = "lint"

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
append_ticket_footer = true
require_ticket = false

[hooks.pre-push]
enabled = false
`,
	"php-node": `[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.ecs]
command = "vendor/bin/ecs"
args = ["check", "--fix"]
type = "php"
extensions = [".php"]
restage = true
group = "lint"

[hooks.pre-commit.tools.prettier]
command = "prettier"
args = ["--write", "--ignore-unknown"]
type = "node"
extensions = [".js", ".ts", ".json", ".md", ".yml", ".yaml"]
restage = true
group = "format"

[hooks.pre-commit.tools.eslint]
command = "eslint"
args = ["--fix", "--cache", "--no-warn-ignored"]
type = "node"
extensions = [".js", ".jsx", ".ts", ".tsx", ".vue"]
restage = true
group = "lint"

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
append_ticket_footer = true
require_ticket = false

[hooks.pre-push]
enabled = false
`,
	"go": `[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.gofmt]
command = "gofmt"
args = ["-w"]
type = "system"
extensions = [".go"]
restage = true
group = "format"

[hooks.pre-commit.tools.govet]
command = "go"
args = ["vet", "./..."]
type = "system"
extensions = [".go"]
pass_files = false
group = "lint"

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
append_ticket_footer = false
require_ticket = false

[hooks.pre-push]
enabled = false
`,
	"minimal": `[hooks.pre-commit]
enabled = true

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true

[hooks.pre-push]
enabled = false
`,
}

// ListPresets returns the sorted preset names.
func ListPresets() []string {
	names := make([]string, 0, len(presets))
	for name := range presets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
