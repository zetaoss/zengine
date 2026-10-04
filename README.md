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
- Runtime configuration (environment variables, files provided at `/files`): `docs/config.md`
- Redis roles (persist / volatile): `docs/redis.md`

## Development Container

The development image is the `dev` stage of the Dockerfile:

```sh
docker build --target dev -t zengine-dev .
```

Inside the container, after cloning, switching branches, or changing
dependencies, synchronize the checkout-specific dependencies without changing
Git state:

```sh
./hack/dev-sync
```

The command reuses pnpm and Go caches under `tmp/` (`tmp/.pnpm-store`,
`tmp/.runtime-cache/`). Database migrations are not run automatically; review
pending migrations and run `ctl migrate` when appropriate.

## MediaWiki Extensions

`mw/extensions.yaml` is the single list of MediaWiki extensions, in two sections (`bundled`, `extra`): which are
enabled, where they come from (repo and tag), and their
configuration. The image installs the enabled ones, and
`ExtensionSettings.php` (`wfLoadExtension` + configuration) is generated from
it. See the header of that file for the fields.

```sh
make extension-settings   # print the generated ExtensionSettings.php
make composer-lock        # after adding/removing installed extensions; commit the lock
```
