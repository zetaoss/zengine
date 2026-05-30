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
- **EditBot**: AI-powered wiki editing agent.
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

- Use standard Tailwind color utilities in markup (for example, `text-slate-700`, `bg-slate-800`).
- Tailwind color utilities automatically respond to dark mode in this project, including `text-white` and `text-black`.
- Do not use `dark:` variants in classes.
- Dark mode colors are applied automatically by the existing theme/token CSS.
- If you need the exact same color in both light and dark mode, use hard-coded app variables via arbitrary values (for example, `bg-(--raw-slate-700) text-(--raw-white)`).
- `svelte/src/shared/assets/appcolor-mw.css` contains MediaWiki semantic tokens, but using those tokens directly is generally not recommended.

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

## Runtime Routing (Nginx Front)

- Nginx listens on `:80` as the public entrypoint.
- `/api/*` -> Go app routes.
- `/auth/*` -> Go app routes.
- `/` -> proxied to Go app (`127.0.0.1:8080`).
- `/api/stat/*` -> served by Go app stat handlers.
- `/wiki/*`, `/w/*` -> MediaWiki rules.

## Subsystem Docs

- Go backend details (including EditBot and write-request specifics): `goapp/AGENTS.md`
