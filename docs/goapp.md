# GoApp Development and Operations Guide

This guide describes the structure of the `goapp` Go backend, local development, HTTP runtime, and Asynq task processing.

## Code structure

| Role | Path |
| --- | --- |
| Process and CLI entry points | `goapp/cmd/{server,worker,scheduler,tool}` |
| API route registry | `goapp/server/routes.go` |
| API and authentication handlers | `goapp/server/handlers/**` |
| Router and middleware | `goapp/server/router/**` |
| Services | `goapp/services/**` |
| Background tasks | `goapp/tasks/**` |
| Models | `goapp/models/**` |
| LLM chains | `goapp/prompts/**` ([chain.md](chain.md)) |
| Task registry and worker | `goapp/worker/**` |
| Database, configuration, and task context | `goapp/app/**` |

HTTP requests generally flow from handler to service or task to model. GORM is used for database operations.

## Processes and deployment topology

```text
server     Any number
worker     Any number
scheduler  Exactly one
```

```bash
cd /app/goapp
go run ./cmd/server
go run ./cmd/worker
go run ./cmd/scheduler
```

- The server handles HTTP requests and enqueues tasks, so it can scale horizontally.
- Asynq coordinates task claims through Redis, so workers can scale horizontally.
- Running multiple schedulers enqueues the same cron tasks from each instance. Keep the deployed scheduler at exactly one replica, and do not combine it with the worker deployment.
- If the scheduler stops, already-enqueued tasks continue to run, but no cron tasks are created while it is down.

The Dockerfile builds `server`, `worker`, `scheduler`, and `tool` binaries into `/app/bin` in the image. Kubernetes selects the process with each container's `command`. The scheduler deployment must run exactly one replica of `/app/bin/scheduler`.

## Development runtime

In development, supervisor manages the three Go processes through Air.

| Process | Air config | Log |
| --- | --- | --- |
| `goserver` | `/app/goapp/.air.server.toml` | `/app/tmp/goserver.log` |
| `goworker` | `/app/goapp/.air.worker.toml` | `/app/tmp/goworker.log` |
| `goscheduler` | `/app/goapp/.air.scheduler.toml` | `/app/tmp/goscheduler.log` |

- All processes use `/app/goapp` as their working directory.
- Do not create separate Air configs under `cmd/server`, `cmd/worker`, or `cmd/scheduler`.
- If a change does not take effect, check the supervisor command, working directory, watched paths in the log, and the latest `building...` entry.

Inspect routes with:

```bash
go run ./cmd/tool routes
```

## HTTP runtime

- Nginx on `:80` is the public entry point.
- `/`, `/api/*`, and `/auth/*` are forwarded to GoApp.
- `/wiki/*` and `/w/*` are served by the MediaWiki stack.
- In dev mode, frontend requests are proxied to Vite at `http://127.0.0.1:5173`.
- In production, `/app/svelte/dist` is served and runtime settings are injected into `window.ZCONF` in `index.html`.

## Middleware and access control

Use the `Router` middleware factory to keep `routes.go` declarative.

- `r.WithUser()`: Adds user information to the context if the user is logged in.
- `r.User()`: Requires a logged-in user.
- `r.Unblocked()`: Requires a logged-in user who is not blocked.
- `r.Sysop()`: Allows only system administrators.
- `r.Internal()`: Allows only internal calls with a valid signature.
- `r.Owner(model)`: Allows only the resource owner.
- `r.OwnerOrSysop(model)`: Allows the owner or a system administrator.

## Main subsystems

### AIEdit

- Route: `/api/ai-edit*` in `goapp/server/routes.go`
- Handler: `goapp/server/handlers/api/aiedit/aiedit.go`
- Model: `goapp/models/ai_edit.go`
- Tasks: `goapp/tasks/aiedit/**`
- Active phases include `Generating` and `Retrying`.
- A single page-level timer refreshes the list and stops when the user leaves the list tab.

`POST /api/ai-edit` renders the request on the `doc-create` or `doc-update` chain on the server (`tasks/aiedit/request.go`, [chain.md](chain.md#chains-and-tasks)): the page, an AI edit prompt template (`prompt_id`) or none, the page as it stands in the editor (`wikitext`, for update) and the run form's values (`user_inputs`). The task stores the template, the chain version, the filled variables (`inputs`) and the rendered messages (`llm_messages`, with `llm_input` as their text); `POST /api/ai-edit/preview` renders the same without storing. A skin loaded before this sends `llm_input`, run as one user message. The template is not edited before a run; the run form only adds values. The `ai-edit` task sends the stored messages with the template's model (`aiedit_prompts.model`), else the chain's model (chains tab, else `chains.docCreate` or `docUpdate`; [chain.md](chain.md#screen-and-api)), so bob does not fall back to another model; a failed call goes to the task's retries and is recorded in `llm_calls`. Request types are `create` (a page that does not exist yet) and `update` (rewrite an existing one), in `aiedit_tasks` and `aiedit_prompts` (`models.AIEditRequestCreate`, `AIEditRequestUpdate`). `update` was called `edit` until v0.9.60; the API still accepts `edit` from skins loaded before the rename. AIEdit prompts (`/api/ai-prompts*`) are sysop-only (`r.Sysop()`). The Activity tab calls the MediaWiki API directly through `mwapi` in Svelte. Contribution flags can be array values such as `"new"` and `"top"`.

### Write Request

- Frontend: `svelte/src/routes/tool/write-request/**`
- Route: `/api/write-request/*` in `goapp/server/routes.go`
- Handler: `goapp/server/handlers/api/writerequest/writerequest.go`
- Tasks: `goapp/tasks/writerequest/**`

### Common Report

- Frontend: `svelte/src/routes/tool/common-report/**`
- Route: `/api/common-report*` in `goapp/server/routes.go`
- Handler: `goapp/server/handlers/api/commonreport/commonreport.go`
- Models: `goapp/models/common_report.go`, `goapp/models/common_report_item.go`
- Tasks: `goapp/tasks/commonreport/**`

### Forum

- Frontend: `svelte/src/routes/forum/**`
- Routes: `/api/posts*`, `/api/posts/{post}/replies*` in `goapp/server/routes.go`
- Handlers: `goapp/server/handlers/api/post/post.go`, `goapp/server/handlers/api/reply/reply.go`
- Models: `goapp/models/forum_post.go`, `goapp/models/forum_reply.go`

### LLM service

- Service: `goapp/services/llmsvc/llmsvc.go`
- Client: `goapp/services/llmsvc/client/client.go`
- Configuration: `API.BobEndpoint` in `goapp/app/config/config.go`; the client calls `BOB_ENDPOINT/aigate/v1/chat/completions`
- Rate limits: a `429` from aigate is retried once after its `Retry-After` (5 seconds when absent); a delay over 30 seconds, such as a daily quota, fails at once. Errors carry aigate's `error` text.
- Call history: `llmsvc.RunStep` passes each model call to the service's recorder; `llmsvc.DBRecorder` stores it in `llm_calls` (`goapp/app/database/migrations/202610250002_llm_calls.up.sql`: chain and version, step, requested and answering model, ok, error, tokens, elapsed time). A response that does not parse is a failed call. It is zengine's empirical record of which models work, kept across bob restarts; the model tab shows it in the same row as bob's usage, after the chains that run on the model: Daily (bob's count for the quota day: calls, limit, remaining, `≈` when bob estimates it, or the time a cooled-down model is back) and Total (calls with the success rate and per-chain counts on hover, latency, first and last success and last failure as ages with the date and time, and the error, on hover). Models with recorded calls that bob no longer lists come last. Rows are not pruned yet.
- Models: `GET /api/ai/models` (`r.Sysop()`, `goapp/server/handlers/api/aimodels/aimodels.go`) returns bob's `/aigate/v1/models?all` and `/aigate/v1/usage` (null when bob lacks it); the AI tool menu's model tab (`svelte/src/routes/tool/ai/models/**`) shows them per provider

### Chains

- Definitions: `goapp/prompts/**`; runner: `llmsvc.RunStep`
- Routes: `GET /api/ai/chains`, `PUT /api/ai/chains/{name}/model` (`r.Sysop()`, [chain.md](chain.md#screen-and-api))
- Frontend: `svelte/src/routes/tool/ai/prompts/**`
- Guide: [chain.md](chain.md)

### Document Evaluation

- Routes: `/api/evaluations*` (`r.Sysop()`) in `goapp/server/routes.go`
- Frontend: `svelte/src/routes/tool/ai/eval/**` (AI menu, sysop-only)
- Handler: `goapp/server/handlers/api/doceval/doceval.go`
- Service: `goapp/services/doceval/doceval.go`
- Model: `goapp/models/doc_evaluation.go`
- Chain: `doc-eval` (`rubric` → `judge`) in `goapp/prompts/chains/doc-eval.yaml`
- Guide: [doc-eval.md](doc-eval.md)

## Asynq task processing

GoApp background tasks use Asynq and Redis. The old custom queue data structures and `Job` interface are not used; there is no Redis migration or compatibility layer.

```text
HTTP server / task handler ---> asynq.Client ---> Redis ---> asynq.Server
Asynq Scheduler -------------> asynq task ----^                |
                                                              v
                                                    Registry -> XxxTask.Execute
```

| Role | Code |
| --- | --- |
| Task implementation | `goapp/tasks/**` |
| Database/configuration/enqueue context | `goapp/app/taskctx/taskctx.go`, `goapp/app/appctx/context.go` |
| Type, timeout, retry, queue, and cron catalog | `goapp/worker/registry/registry.go` |
| Worker lifecycle | `goapp/worker/worker.go` |
| Scheduler lifecycle | `goapp/worker/scheduler/scheduler.go` |
| Inspect, run, and flush CLI | `goapp/cmd/tool/**` |

The server and worker use `asynq.Client` and `asynq.Server` directly. There is no separate queue-client wrapper, request/result adapter, or task-factory package. The registry connects tasks to `ServeMux` and manages static specifications shared by the scheduler and CLI.

### Defining and registering a task

Define each unit of work as an `XxxTask` in `goapp/tasks/<package>`. There are no `Name`, `Run`, or `Timeout` methods and no `job.Result` layer.

```go
type ExampleTask struct{}

func (t *ExampleTask) Execute(
    ctx context.Context,
    taskCtx taskctx.Context,
    payload ExamplePayload,
) (app.H, error)
```

The registry's generic adapter decodes the Asynq JSON payload into its concrete type. Invalid JSON is archived with `asynq.SkipRetry`. Execution errors, panics, and timeouts follow the Asynq retry policy. The returned `app.H` is not stored for asynchronous execution; it is used only as output for direct `tool` execution.

To add a task:

1. Implement `XxxTask`, its payload, and `Execute` in `goapp/tasks/<package>`.
2. Register its type, timeout, queue, retry policy, and cron schedule.
3. If external code must enqueue it, define `asynq.NewTask` and its options in a package helper.
4. Ensure idempotency with database claims, unique constraints, upserts, or similar mechanisms.
5. Run `go test ./...` and `go vet ./...`.

Task type strings and JSON payloads are protocols between deployed versions, so maintain compatibility with tasks enqueued by earlier versions.

### Enqueueing and status

Application enqueue helpers use the official Asynq API and options directly.

```go
task := asynq.NewTask(taskType, payload)
info, err := taskCtx.EnqueueTask(ctx, task,
    asynq.Queue("default"),
    asynq.MaxRetry(3),
    asynq.Timeout(5*time.Minute),
)
```

The default retry count is 3. `common-report` uses `asynq.Unique(30*time.Minute)` and treats `asynq.ErrDuplicateTask` as a request that has already been handled.

```text
scheduled -> pending -> active -> completed
                         |
                         +-> retry -> pending
                         +-> archived
```

Delivery is at least once, so every handler must tolerate duplicate and concurrent execution.

### Worker concurrency and shutdown

One `asynq.Server` in each worker process consumes all queues.

- `Concurrency: 1` means each worker process runs only one task at a time, regardless of queue.
- `default` and `runbox` each have a weight of 1, so they are selected equally when both have pending tasks.
- Increasing worker replicas increases total concurrency by the same factor.

On SIGINT or SIGTERM, active handlers are drained for up to 30 seconds.

### Registered tasks

| Task type | Timeout | Retry | Schedule | Queue / trigger |
| --- | ---: | ---: | --- | --- |
| `ai-edit` | 10 min | 3 | - | API and nanny |
| `ai-edit-nanny` | 1 min | 3 | `0 * * * *` | `default` |
| `common-report` | 5 min | 3 | - | API and nanny, unique for 30 min |
| `common-report-nanny` | 1 min | 3 | `0 * * * *` | `default` |
| `doc-define-loop` | 2 min | 0 | - | `doc-eval`, reschedules itself |
| `doc-eval-loop` | 5 min | 0 | - | `doc-eval`, reschedules itself |
| `inspire` | 5 sec | 3 | - | Manual |
| `ping-db` | 10 sec | 3 | - | Manual |
| `ping-redis` | 5 sec | 3 | - | Manual |
| `request-matcher` | 5 min | 3 | `15 * * * *` | `default` |
| `request-pruner` | 5 min | 3 | `0 0 * * *` | `default` |
| `runbox` | 2 min | 0 | - | `runbox`, API |
| `runbox-pruner` | 1 min | 3 | `* * * * *` | `default`, every minute |
| `stat-{cf,ga,gsc,mw}-{daily,hourly}` | 5 min | 3 | `5 * * * *` | `default` |
| `system-loop-nanny` | 1 min | 3 | `0 * * * *` | `default`, restores missing loop tasks |

The `daily` statistics tasks also run at 5 minutes past every hour; each task determines the collection time range internally.

### Operations CLI

```bash
cd /app/goapp
go run ./cmd/tool tasks
go run ./cmd/tool tasks --watch
go run ./cmd/tool ai-edit '{"task_id":123}'
go run ./cmd/tool flush active
go run ./cmd/tool flush pending
go run ./cmd/tool flush scheduled
go run ./cmd/tool flush retry
go run ./cmd/tool flush all
```

`flush active` sends cancellation to active tasks. `flush pending`, `flush scheduled`, and `flush retry` archive tasks only in the named state; `flush all` handles all four states. Do not modify Asynq Redis keys directly in application code; use `asynq.Inspector` or an official tool.
