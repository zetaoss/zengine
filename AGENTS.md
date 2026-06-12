# Agent Working Guide (/app)

This guide helps coding agents quickly understand the monorepo and choose where to work.

## Core Mandates

- **Contextual Precedence**: This file and directory-level `AGENTS.md` files contain the source of truth for repository workflows and standards.
- **Engineering Standards**:
  - **Go Backend**: Follow clean architecture (handlers -> services/jobs -> models). Use GORM for DB operations.
  - **Svelte Frontend**:
    - Maintain mobile-first responsive design using Tailwind CSS.
    - Use reactive stores for shared state.
    - **Quality Standards**:
      - **No `any`**: Strictly avoid `any` types. Define proper interfaces or types.
      - **Import Sorting**: Always maintain sorted imports (use `eslint --fix`).
      - **Verification**: Run `npm run check` and `eslint` before finalizing changes.
      - **Svelte Warnings**: Treat `svelte-check` warnings as review items and resolve them (or document why they are intentionally kept).
  - **Shared Code**: `svelte/src/shared/` is the source for utilities used by both main and skin Svelte. Never edit the symlink in `mwz/`.

## Scope

- Use `README.md` for a short project overview.
- Use this file for repo-level navigation and investigation workflow.
- Put subsystem details in directory-level `AGENTS.md` files.

## Codebase Map

- `goapp/`: Go backend application
  - `server`: API server handling HTTP requests, routing, and auth
  - `worker`: Background worker for processing jobs and scheduled tasks
  - `ctl`: Command-line tool for managing jobs (list, exec, flush)
- `mwz/`: MediaWiki customization
  - `mwz/extensions/ZetaExtension/`: MediaWiki extension hooks
  - `mwz/skins/ZetaSkin/`: MediaWiki skin integration
  - `mwz/skins/ZetaSkin/svelte/`: **skin svelte** (Svelte components integrated into the skin)
- `svelte/`: **main svelte** (Frontend routes and UI for the main application)
- `w/`: MediaWiki core (Git-ignored)

## Key Subsystems

- **Common-Report**: Aggregated search results and statistics reporting.
- **AIEdit**: AI-powered wiki editing agent.
- **Forum**: Community discussion board for posts and replies.
- **LLM Service**: Unified interface for LLM integrations.
- **Write-Request**: User-driven requests for wiki content creation.

## First Checks By Task

- API behavior:
  - Route table: `goapp/server/routes.go`
  - Route handlers: `goapp/server/handlers/api/**`
- Auth/permission issues:
  - Go middleware: `goapp/server/middleware/**`
  - Social auth handler: `goapp/server/handlers/auth/social/social.go`
- Frontend route wiring:
  - `svelte/src/routes/**`
- MediaWiki hook side effects:
  - `mwz/extensions/ZetaExtension/extension.json`
  - `mwz/extensions/ZetaExtension/includes/**`

## Standard Route Investigation Workflow

1. Confirm route in `goapp/server/routes.go`.
2. Open the mapped handler in `goapp/server/handlers/api/**`.
3. Verify middleware and authorization chain.
4. Identify DB tables/models used by the handler.
5. Trace frontend callers in `svelte/src/routes/**` and shared utils.

## Frontend Styling Rule

- Prefer the design tokens defined in `svelte/src/shared/assets/appcn-globals.css`.
- Use Tailwind utilities to express those tokens in markup, rather than introducing ad hoc CSS.

## Frontend Shared Responsive Utilities

- Mobile breakpoint source of truth: `svelte/src/shared/utils/screen.ts`
  - `isMdOrLargerStore`: Svelte store (reactive)
  - `isMdOrLarger()`: getter function
- Responsive Styling Rules:
  - Breakpoint follows Tailwind's `md` screen.
  - **Mobile-first**: Write default styles for mobile, use Tailwind `md:` utilities for medium-screen enhancements.
  - **Component Scoping**: Component-specific responsive sizes (e.g. modal dimensions) should use component-local Tailwind classes when practical.
- `mwz/skins/ZetaSkin/svelte/src/shared` is a symlink to `svelte/src/shared`.
  - Shared utils, components, and styles are maintained in **main svelte**'s `shared` folder.
  - Changes are automatically reflected in **skin svelte** via this symlink.

## Frontend Dev Runtime

- In local dev, `supervisor` program `dev2` automatically rebuilds/reloads skin Svelte changes.
- `dev2` must run Vite dev server mode (`pnpm run dev:restart`) to preserve HMR in MediaWiki edit/view integration tests.
- Do not switch `dev2` to watch-build mode (`pnpm run watch`) for routine development.
- For `mwz/skins/ZetaSkin/svelte` edits, do **not** assume manual `vite build` is required before verification.
- In dev mode, `mwz/skins/ZetaSkin/svelte/vite.config.ts` serves `src/main.ts` through Vite; `dist/app.js` and `dist/app.css` exist mainly as compatibility stubs for MediaWiki asset injection, not as the live source of truth.
- If browser behavior looks stale in dev, prefer checking the Vite dev server/HMR path and browser cache behavior before rebuilding `dist/`.
- When debugging skin Svelte, remember that the page can reload through MediaWiki while the actual JS module still comes from the Vite dev server; `dist/` timestamps alone do not prove the runtime bundle in use.

## Runtime Routing (Nginx Front)

- Nginx listens on `:80` as the public entrypoint.
- `/api/*` -> Go app routes.
- `/auth/*` -> Go app routes.
- `/` -> proxied to Go app (`127.0.0.1:8080`).
- `/api/stat/*` -> served by Go app stat handlers.
- `/wiki/*`, `/w/*` -> MediaWiki rules.

## Subsystem Docs

- Go backend details (including AIEdit and write-request specifics): `goapp/AGENTS.md`

## Dev Process Notes

- Go server/worker in this environment are supervised by `supervisor` and run through `air`.
- Canonical Air configs are in `goapp/` root:
  - `.air.server.toml`
  - `.air.worker.toml`
- Avoid ad-hoc Air configs in `cmd/server` or `cmd/worker` for routine development; they can cause partial watch coverage and stale binaries.
- If backend code edits appear unapplied, verify:
  1. `supervisor` program commands point to `/app/goapp/.air.server.toml` and `/app/goapp/.air.worker.toml`.
  2. `directory` is `/app/goapp`.
  3. `/app/tmp/goserver.log` shows `watching server/...` and a fresh `building...` after file edits.
