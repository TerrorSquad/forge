import { defineConfig } from 'vitepress'

const hostname = 'https://terrorsquad.github.io/forge/'

export default defineConfig({
  title: 'forge',
  description:
    'A single-binary git hook runner that runs your linters and formatters inside DDEV or Docker containers automatically. No Node.js. A Husky and lefthook alternative for containerized PHP, Go, and polyglot repos.',
  base: '/forge/',
  lastUpdated: true,
  cleanUrls: true,
  sitemap: { hostname },

  head: [
    ['link', { rel: 'icon', href: '/forge/favicon.ico' }],
    ['meta', { name: 'keywords', content: 'git hooks, git hook runner, DDEV, Docker, pre-commit, commit-msg, Husky alternative, lefthook alternative, conventional commits, PHP, Go, monorepo' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:site_name', content: 'forge' }],
    ['meta', { property: 'og:title', content: 'forge — git hook runner for DDEV, Docker & any project' }],
    ['meta', { property: 'og:description', content: 'Run your linters and formatters inside DDEV or Docker containers automatically. Single Go binary, no Node.js. A Husky/lefthook alternative for containerized repos.' }],
    ['meta', { property: 'og:url', content: hostname }],
    ['meta', { property: 'og:image', content: hostname + 'og-image.png' }],
    ['meta', { property: 'og:image:width', content: '1200' }],
    ['meta', { property: 'og:image:height', content: '630' }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:image', content: hostname + 'og-image.png' }],
    ['meta', { name: 'twitter:title', content: 'forge — git hook runner for DDEV, Docker & any project' }],
    ['meta', { name: 'twitter:description', content: 'Run git hooks inside DDEV/Docker containers automatically. Single binary, no Node.js.' }],
  ],

  themeConfig: {
    logo: '/logo.png',

    nav: [
      { text: 'Guide', link: '/guide/installation' },
      { text: 'Reference', link: '/reference/cli' },
      { text: 'Changelog', link: 'https://github.com/TerrorSquad/forge/blob/main/CHANGELOG.md' },
      {
        text: 'GitHub',
        link: 'https://github.com/TerrorSquad/forge',
      },
    ],

    sidebar: [
      {
        text: 'Getting Started',
        items: [
          { text: 'Installation', link: '/guide/installation' },
          { text: 'Quick Start', link: '/guide/quick-start' },
          { text: 'Migrating from Husky', link: '/guide/migrating' },
        ],
      },
      {
        text: 'Guide',
        items: [
          { text: 'Configuration', link: '/guide/configuration' },
          { text: 'Hooks', link: '/guide/hooks' },
          { text: 'Backends (DDEV / Docker)', link: '/guide/backends' },
          { text: 'Workspace / Monorepo', link: '/guide/workspace' },
          { text: 'Commit-message Policy', link: '/guide/commit-policy' },
        ],
      },
      {
        text: 'Reference',
        items: [
          { text: 'CLI Commands', link: '/reference/cli' },
          { text: 'forge.toml', link: '/reference/config' },
          { text: 'Environment Variables', link: '/reference/env' },
          { text: 'Presets', link: '/reference/presets' },
        ],
      },
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/TerrorSquad/forge' },
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © ' + new Date().getFullYear() + ' TerrorSquad',
    },

    search: {
      provider: 'local',
    },

    editLink: {
      pattern: 'https://github.com/TerrorSquad/forge/edit/main/website/:path',
      text: 'Edit this page on GitHub',
    },
  },
})
