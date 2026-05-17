# zengine

Monorepo for ZetaWiki services.

## Components

- `svelte/`: main frontend application
- `goapp/`: primary HTTP server for `/`, `/api/*`, `/auth/*`
- `laravel/`: legacy services and scheduled jobs
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
- Go backend implementation notes: `goapp/AGENTS.md`
