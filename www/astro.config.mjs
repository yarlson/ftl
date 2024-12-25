// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

// https://astro.build/config
export default defineConfig({
  integrations: [
    starlight({
      title: "FTL",
      logo: {
        light: "./src/assets/ftl-light.svg",
        dark: "./src/assets/ftl-dark.svg",
        replacesTitle: true,
      },
      customCss: [
        "@fontsource-variable/rubik",
        "@fontsource-variable/roboto-mono",
        "@fontsource/ibm-plex-mono/400.css",
        "@fontsource/ibm-plex-mono/400-italic.css",
        "@fontsource/ibm-plex-mono/500.css",
        "@fontsource/ibm-plex-mono/600.css",
        "@fontsource/ibm-plex-mono/700.css",
        "./src/custom.css",
      ],
      social: {
        github: "https://github.com/withastro/starlight",
      },
      sidebar: [
        {
          label: "Guides",
          items: [
            { label: "Getting Started", link: "/guides/getting-started/" },
            { label: "Installation", link: "/guides/installation/" },
            {
              label: "Basic Setup",
              collapsed: true,
              items: [
                {
                  label: "Configuration",
                  link: "/guides/basic-setup/configuration/",
                },
                {
                  label: "Environment",
                  link: "/guides/basic-setup/environment/",
                },
                {
                  label: "Dependencies",
                  link: "/guides/basic-setup/dependencies/",
                },
              ],
            },
            {
              label: "Deployment",
              collapsed: true,
              items: [
                { label: "Setup", link: "/guides/deployment/setup/" },
                {
                  label: "Deploy Process",
                  link: "/guides/deployment/deploy-process/",
                },
                {
                  label: "Health Monitoring",
                  link: "/guides/deployment/health-monitoring/",
                },
              ],
            },
            {
              label: "Advanced",
              collapsed: true,
              items: [
                {
                  label: "Nginx Configuration",
                  link: "/guides/advanced/nginx/",
                },
                { label: "SSL Management", link: "/guides/advanced/ssl/" },
                {
                  label: "Zero Downtime Deployment",
                  link: "/guides/advanced/zero-downtime-deployment/",
                },
              ],
            },
          ],
        },
        {
          label: "Reference",
          collapsed: true,
          items: [
            {
              label: "CLI",
              items: [
                { label: "Commands", link: "/reference/cli/commands/" },
                {
                  label: "Configuration",
                  link: "/reference/cli/configuration/",
                },
              ],
            },
            {
              label: "YAML",
              items: [
                { label: "Schema", link: "/reference/yaml/schema/" },
                { label: "Examples", link: "/reference/yaml/examples/" },
              ],
            },
            { label: "Troubleshooting", link: "/reference/troubleshooting/" },
          ],
        },
        {
          label: "Examples",
          collapsed: true,
          items: [
            {
              label: "Simple Web App",
              items: [
                { label: "Overview", link: "/examples/simple-webapp/" },
                {
                  label: "Configuration",
                  link: "/examples/simple-webapp/configuration/",
                },
              ],
            },
            {
              label: "Database Integration",
              items: [
                { label: "Overview", link: "/examples/database-integration/" },
                {
                  label: "Setup",
                  link: "/examples/database-integration/setup/",
                },
              ],
            },
            {
              label: "Microservices",
              items: [
                { label: "Overview", link: "/examples/microservices/" },
                {
                  label: "Architecture",
                  link: "/examples/microservices/architecture/",
                },
              ],
            },
          ],
        },
      ],
    }),
  ],
});
