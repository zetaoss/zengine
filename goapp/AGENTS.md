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

Worker:

```bash
cd /app/goapp
go run ./cmd/worker
```

## 2. Middleware & Security Policy

### Fluent Middleware Factories
To keep `routes.go` clean and declarative, the `Router` struct provides concise middleware factory methods.

*   **`r.WithUser()`**: Optional auth; populates user info into context if logged in.
*   **`r.User()`**: Strictly requires a valid session.
*   **`r.Unblocked()`**: Requires a non-blocked logged-in user. (Recommended for write operations)
*   **`r.Sysop()`**: Strictly restricted to system administrators.
*   **`r.Internal()`**: Restricted to internal system calls (signature required).
*   **`r.Owner(model)`**: Strictly permits only the resource owner. Useful for update (PUT/POST) operations.
*   **`r.OwnerOrSysop(model)`**: Permits sysops or the resource owner. Recommended for delete operations and moderation tasks. It automatically extracts the table name from the provided `model`.

### Resource Access Guidelines
*   **Edit (Update)**: Recommended to use **`r.Owner(models.SomeModel{})`** to ensure only the original author can modify content.
*   **Delete**: Recommended to use **`r.OwnerOrSysop(models.SomeModel{})`** to allow both the owner and system administrators to manage content.

## EditBot Map
...
### EditBot Prompts
- Routes: `/api/editbot/prompts*` in `server/routes.go`
- Permissions: 
    - Create: `r.Unblocked()`
    - Update: Only Author (via `r.Owner(models.EditbotPrompt{})`)
    - Delete: Author or Sysop (via `r.OwnerOrSysop(models.EditbotPrompt{})`)

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
