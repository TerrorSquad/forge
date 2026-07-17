---
layout: home
title: forge — git hook runner for DDEV, Docker & any project
titleTemplate: false
description: Run your linters and formatters inside DDEV or Docker containers automatically. A single-binary git hook runner with no Node.js — a Husky and lefthook alternative for containerized PHP, Go, and polyglot repos.

hero:
  name: forge
  text: The git hook runner for containerized dev
  tagline: Run your linters and formatters inside DDEV or Docker — automatically, no wrapper scripts. One Go binary, no Node.js, any project.
  actions:
    - theme: brand
      text: Get Started
      link: /guide/installation
    - theme: alt
      text: Backends (DDEV & Docker)
      link: /guide/backends

features:
  - icon: 🐳
    title: Runs hooks inside your container
    details: forge auto-detects DDEV and routes tools through the container — phpstan, ecs, and friends run where they actually live. Point any tool at a Docker container instead.

  - icon: ⚡
    title: Single binary, no Node.js
    details: A self-contained Go binary. Drop it on PATH and go — no runtimes, no package.json, no node_modules. Works in PHP, Go, Python, or mixed repos.

  - icon: 🔒
    title: Commit-message policy built in
    details: Enforce Conventional Commits, append ticket footers from the branch name, require ticket IDs — from config, no separate commitlint install.

  - icon: 📦
    title: Monorepo workspace mode
    details: Configure member paths and forge routes each hook to the right workspace member, using that member's own forge.toml.

  - icon: 🔁
    title: Migrate from Husky
    details: One command converts your .git-hooks.config.json to forge.toml. Keep your workflow, drop the Node dependency.

  - icon: 🛠️
    title: TOML config, editor schema
    details: A readable forge.toml with a JSON schema installed for autocompletion. Presets for node, php, php-node, go, and minimal.
---
