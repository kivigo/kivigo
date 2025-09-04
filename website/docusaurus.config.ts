import { themes as prismThemes } from 'prism-react-renderer';
import type { Config } from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config: Config = {
  title: 'KiviGo',
  tagline: 'Lightweight key-value store library for Go',
  favicon: 'img/favicon.ico',

  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: {
    v4: true, // Improve compatibility with the upcoming Docusaurus v4
  },

  // Set the production url of your site here
  url: 'https://azrod.github.io',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/kivigo/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'azrod', // Usually your GitHub org/user name.
  projectName: 'kivigo', // Usually your repo name.

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            'https://github.com/azrod/kivigo/tree/main/website/',
          versions: {
            current: {
              label: 'current',
            },
          },
          lastVersion: '1.5.0',
          includeCurrentVersion: true
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    // Replace with your project's social card
    image: 'img/kivigo-white.png',
    colorMode: {
      defaultMode: 'dark',
      disableSwitch: true,
      respectPrefersColorScheme: false,
    },
    navbar: {
      title: 'KiviGo',
      logo: {
        alt: 'KiviGo Logo',
        src: 'img/logo-kivigo.png',
      },
      items: [
        {
          type: 'dropdown',
          label: 'Why KiviGo?',
          position: 'left',
          items: [
            {
              label: 'Overview',
              href: '/why/overview',
            },
            {
              label: 'Comparison',
              href: '/why/comparison',
            },
            {
              label: 'Design Decisions',
              href: '/why/design-decisions',
            },
          ],
        },
        {
          type: 'docSidebar',
          sidebarId: 'documentationSidebar',
          position: 'left',
          label: 'Documentation',
        },
        {
          type: 'docsVersionDropdown',
          position: 'right',
          dropdownActiveClassDisabled: true,
        },
        {
          href: 'https://github.com/azrod/kivigo',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Getting Started',
              to: '/docs/getting-started',
            },
            {
              label: 'Backends',
              to: '/docs/backends/overview',
            },
            {
              label: 'Advanced',
              to: '/docs/advanced/health-checks',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'GitHub Issues',
              href: 'https://github.com/azrod/kivigo/issues',
            },
            {
              label: 'Go Reference',
              href: 'https://pkg.go.dev/github.com/azrod/kivigo',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/azrod/kivigo',
            },
            {
              label: 'License',
              href: 'https://github.com/azrod/kivigo/blob/main/LICENSE',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} KiviGo. Built with Docusaurus.`,
    },
    algolia: {
      // The application ID provided by Algolia
      appId: 'FLTTNEGAAP',

      // Public API key: it is safe to commit it
      apiKey: 'b6b4d2b5aa6dbbef9815f72e379efede', // Use the Search API Key, not the Write API Key

      indexName: 'kivigo_v1', // Choose a meaningful name for your index

      // Optional: see doc section below
      contextualSearch: true,

      // Optional: Specify domains where the navigation should occur through window.location instead on history.push
      // externalUrlRegex: 'external\\.com|domain\\.com',

      // Optional: Algolia search parameters
      searchParameters: {},

      // Optional: path for search page that enabled by default (`false` to disable it)
      searchPagePath: 'search',

      // Optional: whether the insights feature is enabled or not on Docsearch
      insights: true,
    },
    prism: {
      theme: prismThemes.gruvboxMaterialDark,
      additionalLanguages: ['go', 'bash', 'yaml', 'json'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
