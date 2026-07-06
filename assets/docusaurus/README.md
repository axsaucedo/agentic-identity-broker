# Website

This website is built using [Docusaurus](https://docusaurus.io/), a modern static website generator.

## Installation

```bash
yarn
```

## Local Development

```bash
yarn start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

## Build

```bash
yarn build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

## Deployment

GitHub Pages deployment is wired for a custom domain and a configurable GitHub org/repo.

Before deploying:

- set `DOCS_SITE_URL=https://<custom-domain>`
- set `DOCS_GITHUB_ORG=<github-org>`
- set `DOCS_GITHUB_REPO=<github-repo>`
- optionally set `DOCS_GITHUB_BRANCH=<default-branch>` (defaults to `main`)

`docs-deploy` derives `build/CNAME` from `DOCS_SITE_URL`, so the custom-domain host only needs to be set in one place.

Then run:

```bash
npm run build
npm run deploy
```
