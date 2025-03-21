import { defineConfig } from "vitepress";
import { withMermaid } from "vitepress-plugin-mermaid";

export default withMermaid(
  defineConfig({
    title: "FTL Documentation",
    description: "Documentation for FTL - Faster Than Light Deployment Tool",
    srcDir: "./docs",

    head: [
      [
        'script',
        {
          defer: 'true',
          'data-domain': 'ftl-deploy.org',
          src: 'https://plausible.io/js/script.outbound-links.js'
        }
      ]
    ],

    markdown: {
      theme: {
        light: "github-light",
        dark: "github-dark",
      },
      languages: ["yaml", "bash", "javascript", "json", "dockerfile"],
    },

    themeConfig: {
      nav: [
        { text: "Home", link: "/" },
        { text: "Getting Started", link: "/getting-started/" },
        { text: "Core Tasks", link: "/core-tasks/" },
        { text: "Configuration", link: "/configuration/" },
        { text: "Guides", link: "/guides/" },
        { text: "Reference", link: "/reference/" },
      ],

      sidebar: {
        "/": [
          {
            text: "Getting Started",
            collapsed: false,
            items: [
              { text: "Introduction", link: "/getting-started/" },
              { text: "Installation", link: "/getting-started/installation" },
              { text: "Configuration", link: "/getting-started/configuration" },
              {
                text: "First Deployment",
                link: "/getting-started/first-deployment",
              },
            ],
          },
          {
            text: "Core Tasks",
            collapsed: false,
            items: [
              { text: "Overview", link: "/core-tasks/" },
              { text: "Server Setup", link: "/core-tasks/server-setup" },
              { text: "Building", link: "/core-tasks/building" },
              { text: "Deployment", link: "/core-tasks/deployment" },
              { text: "Logging", link: "/core-tasks/logging" },
              { text: "Tunneling", link: "/core-tasks/tunneling" },
            ],
          },
          {
            text: "Configuration",
            collapsed: true,
            items: [
              { text: "Overview", link: "/configuration/" },
              {
                text: "Project Settings",
                link: "/configuration/project-settings",
              },
              { text: "Services", link: "/configuration/services" },
              { text: "Dependencies", link: "/configuration/dependencies" },
              { text: "Volumes", link: "/configuration/volumes" },
            ],
          },
          {
            text: "Guides",
            collapsed: true,
            items: [
              { text: "Overview", link: "/guides/" },
              {
                text: "Zero-downtime Deployment",
                link: "/guides/zero-downtime",
              },
              { text: "Health Checks", link: "/guides/health-checks" },
              { text: "SSL Management", link: "/guides/ssl-management" },
            ],
          },
          {
            text: "Reference",
            collapsed: true,
            items: [
              { text: "Overview", link: "/reference/" },
              { text: "CLI Commands", link: "/reference/cli-commands" },
              {
                text: "Configuration File",
                link: "/reference/configuration-file",
              },
              { text: "Environment Variables", link: "/reference/environment" },
              { text: "Troubleshooting", link: "/reference/troubleshooting" },
            ],
          },
        ],
      },

      socialLinks: [{ icon: "github", link: "https://github.com/yarlson/ftl" }],

      footer: {
        message: "Released under the MIT License.",
        copyright: "Copyright © 2024-present FTL Contributors",
      },

      search: {
        provider: "local",
      },

      outline: {
        level: [2, 3],
        label: "On this page",
      },
    },
  }),
);
