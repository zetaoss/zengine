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
go run ./cmd/server
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
- Job: `jobs/editbotjob/editbotjob.go`

### Activity Feed Rule

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

## Common-Report Map

- Frontend: `svelte/src/routes/tool/common-report/**`
- Routes: `/api/common-report*` in `server/routes.go`
- Handler: `server/handlers/api/commonreport/commonreport.go`
- Model: `models/commonreport.go`
- Job: `jobs/commonreportjob/commonreportjob.go`

## Forum (Posts/Replies) Map

- Frontend: `svelte/src/routes/forum/**`
- Routes: `/api/posts*`, `/api/posts/{post}/replies*` in `server/routes.go`
- Handlers: `server/handlers/api/post/post.go`, `server/handlers/api/reply/reply.go`
- Models: `models/forum.go` (Note: ForumPost and ForumReply are defined here)

## LLM Service Map

- Service implementation: `services/llmsvc/llmsvc.go`
- Client: `services/llmsvc/client/client.go`
- Configuration: `app/config/config.go` (via `API.LLMEndpoint`)

## Operational Notes

- In `dev` mode, Go app proxies frontend requests to Vite (`http://127.0.0.1:5173`).
- In `prod` mode, Go app serves `/app/svelte/dist` and injects runtime config into `index.html`.
- Runtime config is exposed as `window.ZCONF`.
