# Agent Working Guide (/app)

This file is for coding agents working in this monorepo.

## Intent

- `README.md` is for project overview.
- `AGENTS.md` is for execution guidance: where to look first, what to verify, and common pitfalls.

## First Places To Check

- API behavior: `laravel/routes/api.php`, then controller in `laravel/app/Http/Controllers/`
- Auth/permissions: `laravel/app/Providers/AuthServiceProvider.php`, `mwauth` middleware usage in routes
- Frontend page wiring: `svelte/src/routes/**`
- MediaWiki hook side effects: `mwz/extensions/ZetaExtension/extension.json` and `includes/**`

## Frontend Styling Guideline

- When editing Svelte UI, prefer existing design tokens over hardcoded colors, checking `svelte/src/shared/assets/appcolor-mw.css` first.

## Route Investigation Workflow

1. Confirm route in `laravel/routes/api.php` or `laravel/routes/web.php`.
2. Open mapped controller action.
3. Check authorization (`Gate::authorize`, middleware).
4. Check DB tables used (`DB::table`, model).
5. Trace frontend caller in `svelte/src/routes` and shared utils.

Useful command:

```bash
cd /app/laravel && php artisan route:list
```

## Write-Request Feature Map

Frontend:

- Page: `svelte/src/routes/tool/write-request/+page.svelte`
- UI: `svelte/src/routes/tool/write-request/WriteRequest.svelte`
- Create modal: `svelte/src/routes/tool/write-request/WriteRequestNew.svelte`

Backend:

- Routes: `/api/write-request/*` in `laravel/routes/api.php`
- Controller: `laravel/app/Http/Controllers/WriteRequestController.php`

MediaWiki integration:

- Hook registration: `mwz/extensions/ZetaExtension/extension.json`
- Mark done on page save: `includes/Binder/BinderHooks.php` -> `WriteRequestService::markDoneIfMatched`
- Search no-match hit count: `includes/WriteRequest/SearchHooks.php`

## Known Gotcha

- `WriteRequestController::update()` exists but no current route binds to it.
