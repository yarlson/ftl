# FTL Documentation Site

Documentation site for [FTL (Faster Than Light Deployment)](https://github.com/yarlson/ftl) built with [Starlight](https://starlight.astro.build) and [Bun](https://bun.sh).

## Quick Start

```bash
# Install dependencies
bun install

# Start dev server
bun dev

# Build for production
bun build

# Preview production build
bun preview
```

## Development

The site uses Starlight's standard directory structure:

```
src/
├── content/
│   └── docs/
│       └── ...mdx files
├── components/
└── assets/
```

## Code Style

We use a combination of Prettier and Biome for formatting:

- Prettier handles Astro and MDX files
- Biome handles JavaScript/TypeScript

Format all files:

```bash
bun format
```

## Contributing

1. Make your changes
2. Run `bun format` before committing
3. Preview your changes with `bun preview`
4. Submit a PR

## Local Development with FTL

To test documentation changes against local FTL changes:

1. Clone FTL repo if you haven't already
2. Run docs site with `bun dev`
3. Make changes to both code and docs
4. Use `bun preview` to verify production build
