# Octopulse

Octopulse is a modern uptime monitoring / status page

Octopulse base on Nuxt.js and [@nuxthub/core](https://nuxt.com/modules/hub), learn more on [hub.nuxt.com](https://hub.nuxt.com/).

## Development

### Setup

Make sure to install the dependencies with pnpm.

```sh
pnpm install
```

Then start development server with `dev`

```sh
pnpm dev
```

## Deploy

### Vercel

On the Vercel dashboard -> Storage, create a Turso SQLite database.

```sh
npm run deploy-vercel
```

Link the newly created resources to your project and re-deploy.

You can also connect using your git repository to leverage Vercel's CI/CD pipeline.

### Cloudflare worker

```sh
npm run deploy-cloudflare
```

Wrangler will create the necessary resources on your Cloudflare account.

You can also connect using your git repository to leverage Cloudflare's CI/CD pipeline.
