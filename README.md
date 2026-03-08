# zengine

Monorepo for the zengine stack (Svelte + Laravel + MediaWiki).

## What Runs Where

- Frontend app: `svelte/` (served at `/`)
- API backend: `laravel/` (served at `/api`)
- Wiki core: `w/` (served at `/w`, `/wiki`)
- Wiki custom code: `mwz/extensions/ZetaExtension/`, `mwz/skins/ZetaSkin/`
- Front router/server: `gohttp/` (port `8080`)

## Repository Map

```text
.
|-- gohttp/                  # Go HTTP server (entry routing)
|-- laravel/                 # API server
|-- mwz/extensions/ZetaExtension/
|-- mwz/skins/ZetaSkin/
|-- svelte/                  # Main frontend app
`-- w/                       # MediaWiki core
```

## Laravel Routing Overview

Route files:

- `laravel/routes/api.php` -> `/api/*`
- `laravel/routes/web.php` -> `/auth/*`
- `laravel/routes/console.php` -> scheduled jobs

Main API domains:

- `/api/comments/*`
- `/api/common-report/*`
- `/api/dash/*`
- `/api/me*`
- `/api/onelines*`
- `/api/posts*`
- `/api/reactions/page*`
- `/api/runbox*`
- `/api/user/*`
- `/api/write-request/*`

## Write Request (Key Endpoints)

- `GET /api/write-request/count`
- `GET /api/write-request/todo`
- `GET /api/write-request/todo-top`
- `GET /api/write-request/done`
- `POST /api/write-request` (`mwauth`)
- `POST /api/write-request/{writeRequest}/recommend` (`mwauth`)
- `DELETE /api/write-request/{writeRequest}` (`mwauth`)

Note: `WriteRequestController::update()` exists but is not route-bound.
