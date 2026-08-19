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

## Kubernetes Development Workspace

After cloning or switching branches in the dev3 workspace, synchronize the
checkout-specific dependencies without changing Git state:

```sh
./hack/dev-sync
```

The command reuses pnpm and Go caches under `.runtime-cache/`. Database
migrations are reported but are never run automatically.
