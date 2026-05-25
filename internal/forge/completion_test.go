package forge

import (
	"strings"
	"testing"
)

func TestGenerateCompletion_Bash(t *testing.T) {
	script, err := GenerateCompletion("bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{
		"forge", "pre-commit", "commit-msg", "run", "init", "--preset", "--all-files",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("bash completion missing %q", want)
		}
	}
}

func TestGenerateCompletion_Zsh(t *testing.T) {
	script, err := GenerateCompletion("zsh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{
		"#compdef forge", "pre-commit", "commit-msg", "--all-files", "--preset",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("zsh completion missing %q", want)
		}
	}
}

func TestGenerateCompletion_Fish(t *testing.T) {
	script, err := GenerateCompletion("fish")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{
		"pre-commit", "commit-msg", "all-files", "bash", "zsh", "fish",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("fish completion missing %q", want)
		}
	}
}

func TestGenerateCompletion_CaseInsensitive(t *testing.T) {
	for _, shell := range []string{"BASH", "ZSH", "Fish"} {
		_, err := GenerateCompletion(shell)
		if err != nil {
			t.Errorf("GenerateCompletion(%q) should be case-insensitive, got: %v", shell, err)
		}
	}
}

func TestGenerateCompletion_UnknownShell(t *testing.T) {
	_, err := GenerateCompletion("powershell")
	if err == nil {
		t.Error("expected error for unknown shell")
	}
	if !strings.Contains(err.Error(), "powershell") {
		t.Errorf("error should name the bad shell, got: %v", err)
	}
}

func TestGenerateCompletion_ContainsAllPresets(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		script, _ := GenerateCompletion(shell)
		for _, p := range ListPresets() {
			if !strings.Contains(script, p) {
				t.Errorf("%s completion missing preset %q", shell, p)
			}
		}
	}
}
