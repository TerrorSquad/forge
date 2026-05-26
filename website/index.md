---
layout: home

hero:
  name: forge
  text: Policy-driven git hook runner
  tagline: Fast, portable, and no Node.js required. One binary. Any project.
  actions:
    - theme: brand
      text: Get Started
      link: /guide/installation
    - theme: alt
      text: View on GitHub
      link: https://github.com/TerrorSquad/forge

features:
  - icon: ⚡
    title: Single binary
    details: A single self-contained Go binary. Drop it on PATH and go. No runtimes, no package managers.

  - icon: 🛠️
    title: Works in any project
    details: PHP, Go, Python, Node — forge doesn't care. Use it wherever you have a git repo.

  - icon: 🐳
    title: DDEV-aware
    details: Run linters and formatters inside your DDEV container automatically, without wrapper scripts.

  - icon: 📦
    title: Monorepo workspace mode
    details: Configure member paths and forge routes hooks to the right workspace member automatically.

  - icon: 🔒
    title: Commit-message policy
    details: Enforce Conventional Commits, append JIRA ticket footers, require ticket IDs — all from config.

  - icon: 🔁
    title: Migrate from Husky
    details: One command converts your .git-hooks.config.json to forge.toml. Keep your workflow, drop Node.
---
