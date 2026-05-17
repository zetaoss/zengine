# GoApp Agent Guide (/app/goapp)

Execution notes for agents working in the Go backend.

## Primary Entry Points

- Route registry: `server/routes.go`
- API handlers: `server/handlers/api/**`
- Auth handlers: `server/handlers/auth/**`
- Middleware: `server/middleware/**`
- Models: `models/**`
- Jobs: `jobs/**`

## Run

```bash
cd /app/goapp
go run ./cmd/server --dev=true
```

```bash
cd /app/goapp
go run ./cmd/server
```

Worker:

```bash
cd /app/goapp
go run ./cmd/worker
```

## EditBot Map

### Frontend Integration

- List page: `svelte/src/routes/tool/editbot/+page.svelte`
- Detail page: `svelte/src/routes/tool/editbot/[id]/+page.svelte`
- MediaWiki API client: `svelte/src/lib/utils/mwapi.ts`

### Backend Integration

- Routes: `/api/editbot*` in `server/routes.go`
- Handler: `server/handlers/api/editbot/editbot.go`
- Model: `models/editbot.go`
- Job: `jobs/editbot/edit_bot_job.go`

### Activity Feed Rule

- `/api/editbot/activity` is removed.
- The activity tab must call MediaWiki API directly from Svelte via `mwapi`.
- MediaWiki contribution flags may come as array values (for example `"new"`, `"top"`) rather than object booleans.

### Queue/Phase Notes

- Active phases include:
  - `Generating`
  - `Publishing`
  - `RetryingGenerate`
  - `RetryingPublish`
- List-page auto-refresh should be a single page-level timer, paused when leaving the list tab.

## Write-Request Map

- Frontend page: `svelte/src/routes/tool/write-request/+page.svelte`
- Main UI: `svelte/src/routes/tool/write-request/WriteRequest.svelte`
- Create modal: `svelte/src/routes/tool/write-request/WriteRequestNew.svelte`
- Routes: `/api/write-request/*` in `server/routes.go`
- Handler: `server/handlers/api/writerequest/writerequest.go`

Known behavior:

- Write-request update exists in code but is not exposed as a route.

## Operational Notes

- In `dev` mode, Go app proxies frontend requests to Vite (`http://127.0.0.1:5173`).
- In `prod` mode, Go app serves `/app/svelte/dist` and injects runtime config into `index.html`.
- Runtime config is exposed as `window.ZCONF`.
