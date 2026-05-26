import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'forge',
  description: 'Policy-driven git hook runner — fast, portable, no Node.js required.',
  base: '/forge/',

  head: [
    ['link', { rel: 'icon', href: '/forge/favicon.ico' }],
  ],

  themeConfig: {
    logo: '/logo.svg',

    nav: [
      { text: 'Guide', link: '/guide/installation' },
      { text: 'Reference', link: '/reference/cli' },
      { text: 'Changelog', link: 'https://github.com/TerrorSquad/forge/blob/master/CHANGELOG.md' },
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
          { text: 'Backends (DDEV)', link: '/guide/backends' },
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
      pattern: 'https://github.com/TerrorSquad/forge/edit/master/website/:path',
      text: 'Edit this page on GitHub',
    },
  },
})
