# forge demo tape — generates demo.gif
# Requirements: vhs (https://github.com/charmbracelet/vhs)
# Run: vhs demo.tape

Output demo.gif
Output demo.mp4

Set FontSize 14
Set Width 900
Set Height 550
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms
Set Padding 20

# ── Initialise a new Go project ──────────────────────────────────────────────
Type "cd /tmp && mkdir demo-repo && cd demo-repo && git init"
Enter
Sleep 400ms

Type "forge init --preset go"
Enter
Sleep 600ms

# ── Show the generated config ─────────────────────────────────────────────────
Type "cat forge.toml"
Enter
Sleep 800ms

# ── Install hooks ─────────────────────────────────────────────────────────────
Type "forge install"
Enter
Sleep 600ms

# ── Doctor check ──────────────────────────────────────────────────────────────
Type "forge doctor"
Enter
Sleep 800ms

# ── Simulate a commit ─────────────────────────────────────────────────────────
Type "echo 'package main' > main.go && git add main.go"
Enter
Sleep 400ms

Type 'git commit -m "feat: initial commit"'
Enter
Sleep 1200ms

# ── Skip a tool ───────────────────────────────────────────────────────────────
Type 'SKIP_GOFMT=1 git commit --allow-empty -m "chore: skip formatter"'
Enter
Sleep 1000ms

Sleep 2s
