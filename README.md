# zengine

Monorepo for ZetaWiki services.

## Components

- `svelte/`: main frontend application
- `goapp/`: primary HTTP server for `/`, `/api/*`, `/auth/*`, background and scheduled tasks
- `w/`: MediaWiki core
- `mwz/extensions/ZetaExtension/`: custom MediaWiki extension
- `mwz/skins/ZetaSkin/`: custom MediaWiki skin

## Routing (High Level)

- `/` -> Go app (and frontend)
- `/api/*` -> Go app API routes
- `/auth/*` -> Go app auth routes
- `/wiki/*`, `/w/*` -> MediaWiki stack

## Developer Docs

- Agent execution guide: `AGENTS.md`
- GoApp development and task system: `docs/goapp.md`

## Development Container

Inside the development container (`ghcr.io/zetaoss/zdev`), after cloning,
switching branches, or changing dependencies, synchronize the
checkout-specific dependencies without changing Git state:

```sh
./hack/dev-sync
```

The command reuses pnpm and Go caches under `tmp/` (`tmp/.pnpm-store`,
`tmp/.runtime-cache/`). Database migrations are not run automatically; review
pending migrations and run `ctl migrate` when appropriate.
