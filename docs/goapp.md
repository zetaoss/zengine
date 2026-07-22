# GoApp 개발 및 운영 가이드

`goapp` Go 백엔드의 구조, 로컬 실행, HTTP runtime과 Asynq task 수행 체계를 정리한다.

## 코드 구조

| 역할 | 경로 |
| --- | --- |
| Process 및 CLI entrypoint | `goapp/cmd/{server,worker,scheduler,ctl,adm}` |
| API route registry | `goapp/server/routes.go` |
| API·인증 handler | `goapp/server/handlers/**` |
| Router와 middleware | `goapp/server/router/**` |
| Service | `goapp/services/**` |
| Background task | `goapp/tasks/**` |
| Model | `goapp/models/**` |
| Task registry와 worker | `goapp/worker/**` |
| DB/config/task context | `goapp/app/**` |

HTTP 요청은 일반적으로 handler -> service/task -> model 순서로 처리하며 DB 작업에는 GORM을 사용한다.

## Process와 배포 토폴로지

```text
server     N개 가능
worker     N개 가능
scheduler  정확히 1개
```

```bash
cd /app/goapp
go run ./cmd/server
go run ./cmd/worker
go run ./cmd/scheduler
```

- Server는 HTTP 요청을 처리하고 task를 enqueue하므로 수평 확장할 수 있다.
- Worker는 Asynq가 Redis에서 task claim을 조정하므로 수평 확장할 수 있다.
- Scheduler를 둘 이상 실행하면 각 instance가 동일 cron task를 enqueue한다. 배포 replica는 반드시 1로 고정하고 worker deployment에 합치지 않는다.
- Scheduler가 중단되어도 이미 enqueue된 task는 계속 처리되지만 중단 중 cron task는 생성되지 않는다.

Dockerfile은 `server`, `worker`, `scheduler` 바이너리를 빌드해 image에 포함하기만 한다. 실제 process는 Kubernetes의 container `command`로 선택하며, scheduler deployment는 `/app/goapp/scheduler`를 정확히 1 replica로 실행한다.

## 개발 Runtime

개발 환경에서는 supervisor가 Air를 통해 세 Go process를 관리한다.

| Process | Air config | Log |
| --- | --- | --- |
| `goserver` | `/app/goapp/.air.server.toml` | `/app/tmp/goserver.log` |
| `goworker` | `/app/goapp/.air.worker.toml` | `/app/tmp/goworker.log` |
| `goscheduler` | `/app/goapp/.air.scheduler.toml` | `/app/tmp/goscheduler.log` |

- Working directory는 모두 `/app/goapp`이다.
- `cmd/server`, `cmd/worker`, `cmd/scheduler` 아래에 별도 Air config를 만들지 않는다.
- 변경이 반영되지 않으면 supervisor command, working directory, log의 감시 경로와 최근 `building...` 기록을 확인한다.

Route는 다음 명령으로 확인한다.

```bash
go run ./cmd/ctl routes
```

## HTTP Runtime

- Nginx `:80`이 public entrypoint다.
- `/`, `/api/*`, `/auth/*`는 GoApp으로 전달된다.
- `/wiki/*`, `/w/*`는 MediaWiki stack으로 전달된다.
- Dev mode에서는 frontend 요청을 Vite `http://127.0.0.1:5173`으로 proxy한다.
- Prod mode에서는 `/app/svelte/dist`를 제공하고 runtime 설정을 `index.html`의 `window.ZCONF`로 주입한다.

## Middleware와 접근 제어

`Router`의 middleware factory를 사용해 `routes.go`를 선언적으로 유지한다.

- `r.WithUser()`: 로그인했다면 사용자 정보를 context에 추가한다.
- `r.User()`: 로그인을 요구한다.
- `r.Unblocked()`: 차단되지 않은 로그인 사용자를 요구한다.
- `r.Sysop()`: 시스템 관리자만 허용한다.
- `r.Internal()`: 유효한 signature를 가진 내부 호출만 허용한다.
- `r.Owner(model)`: resource 작성자만 허용한다.
- `r.OwnerOrSysop(model)`: 작성자 또는 관리자를 허용한다.

## 주요 Subsystem

### AIEdit

- Route: `goapp/server/routes.go`의 `/api/ai-edit*`
- Handler: `goapp/server/handlers/api/aiedit/aiedit.go`
- Model: `goapp/models/ai_edit.go`
- Task: `goapp/tasks/aiedit/**`
- 활성 phase에는 `Generating`, `Retrying`이 포함된다.
- 목록 자동 새로고침은 page-level timer 하나를 사용하고 목록 tab을 벗어나면 중지한다.

AIEdit prompt는 생성에 `r.Unblocked()`, 수정에 `r.Owner(models.AIEditPrompt{})`, 삭제에 `r.OwnerOrSysop(models.AIEditPrompt{})`를 사용한다. Activity tab은 Svelte에서 `mwapi`로 MediaWiki API를 직접 호출하며 contribution flag는 `"new"`, `"top"` 같은 배열 값일 수 있다.

### Write Request

- Frontend: `svelte/src/routes/tool/write-request/**`
- Route: `goapp/server/routes.go`의 `/api/write-request/*`
- Handler: `goapp/server/handlers/api/writerequest/writerequest.go`
- Task: `goapp/tasks/writerequest/**`

### Common Report

- Frontend: `svelte/src/routes/tool/common-report/**`
- Route: `goapp/server/routes.go`의 `/api/common-report*`
- Handler: `goapp/server/handlers/api/commonreport/commonreport.go`
- Model: `goapp/models/common_report.go`, `goapp/models/common_report_item.go`
- Task: `goapp/tasks/commonreport/**`

### Forum

- Frontend: `svelte/src/routes/forum/**`
- Route: `goapp/server/routes.go`의 `/api/posts*`, `/api/posts/{post}/replies*`
- Handler: `goapp/server/handlers/api/post/post.go`, `goapp/server/handlers/api/reply/reply.go`
- Model: `goapp/models/forum_post.go`, `goapp/models/forum_reply.go`

### LLM Service

- Service: `goapp/services/llmsvc/llmsvc.go`
- Client: `goapp/services/llmsvc/client/client.go`
- 설정: `goapp/app/config/config.go`의 `API.LLMEndpoint`

## Asynq Task 수행 체계

GoApp의 background task는 Asynq와 Redis를 사용한다. 기존 자체 queue 자료구조와 `Job` interface는 사용하지 않으며 Redis migration이나 호환 계층도 두지 않는다.

```text
HTTP server / task handler ---> asynq.Client ---> Redis ---> asynq.Server
Asynq Scheduler -------------> asynq task ----^                |
                                                              v
                                                    Registry -> XxxTask.Execute
```

| 역할 | 코드 |
| --- | --- |
| Task 구현 | `goapp/tasks/**` |
| DB/config/enqueue context | `goapp/app/taskctx/taskctx.go`, `goapp/app/appctx/context.go` |
| Type, timeout, retry, queue, cron catalog | `goapp/worker/registry/registry.go` |
| Worker lifecycle | `goapp/worker/worker.go` |
| Scheduler lifecycle | `goapp/worker/scheduler/scheduler.go` |
| 조회·직접 실행·flush CLI | `goapp/cmd/ctl/**` |

Server와 worker는 `asynq.Client`와 `asynq.Server`를 직접 사용한다. 별도의 queue client wrapper, request/result adapter, task factory package는 없다. Registry는 task를 `ServeMux`에 연결하고 scheduler와 CLI가 공유하는 정적 spec을 관리한다.

### Task 작성과 등록

업무 단위는 `goapp/tasks/<package>`의 `XxxTask`로 정의한다. `Name`, `Run`, `Timeout` method나 `job.Result` 계층은 없다.

```go
type ExampleTask struct{}

func (t *ExampleTask) Execute(
    ctx context.Context,
    taskCtx taskctx.Context,
    payload ExamplePayload,
) (app.H, error)
```

Registry의 generic adapter가 Asynq JSON payload를 구체 payload type으로 decode한다. 잘못된 JSON은 `asynq.SkipRetry`로 archive한다. 실행 error, panic, timeout은 Asynq retry 정책을 따른다. 반환 `app.H`는 비동기 실행에서 저장하지 않고 `ctl` 직접 실행 출력에만 사용한다.

새 task 추가 절차:

1. `goapp/tasks/<package>`에 `XxxTask`, payload, `Execute`를 구현한다.
2. Registry에 type, timeout, queue, retry, cron을 등록한다.
3. 외부 enqueue가 필요하면 package helper에서 `asynq.NewTask`와 옵션을 정의한다.
4. DB claim, unique constraint, upsert 등으로 멱등성을 확보한다.
5. `go test ./...`와 `go vet ./...`를 실행한다.

Task type 문자열과 JSON payload는 배포 version 사이의 protocol이므로 이전 version이 enqueue한 task와의 호환성을 고려한다.

### Enqueue와 상태

Application enqueue helper는 공식 Asynq API와 option을 직접 사용한다.

```go
task := asynq.NewTask(taskType, payload)
info, err := taskCtx.EnqueueTask(ctx, task,
    asynq.Queue("default"),
    asynq.MaxRetry(3),
    asynq.Timeout(5*time.Minute),
)
```

기본 retry는 3회다. `common-report`는 `asynq.Unique(30*time.Minute)`를 사용하며 `asynq.ErrDuplicateTask`를 이미 요청된 것으로 보고 성공 처리한다.

```text
scheduled -> pending -> active -> completed
                         |
                         +-> retry -> pending
                         +-> archived
```

전달은 at-least-once이므로 모든 handler는 중복·동시 실행에 안전해야 한다.

### Worker 동시성과 종료

Worker process는 하나의 `asynq.Server`로 모든 queue를 소비한다.

- `Concurrency: 1`이므로 한 worker process에서는 queue와 무관하게 task를 한 번에 하나만 실행한다.
- `default`와 `runbox`의 weight는 각각 1이며, 둘 다 대기 중이면 동일한 비율로 선택된다.
- Worker replica가 늘면 전체 concurrency도 replica 수만큼 증가한다.

SIGINT/SIGTERM 시 active handler를 최대 30초 drain한다.

### 등록된 Task

| Task type | Timeout | Retry | Schedule | Queue / trigger |
| --- | ---: | ---: | --- | --- |
| `ai-edit` | 10분 | 3 | - | API 및 nanny |
| `ai-edit-nanny` | 1분 | 3 | `0 * * * *` | `default` |
| `common-report` | 5분 | 3 | - | API 및 nanny, unique 30분 |
| `common-report-nanny` | 1분 | 3 | `0 * * * *` | `default` |
| `inspire` | 5초 | 3 | - | 수동 |
| `ping-db` | 10초 | 3 | - | 수동 |
| `ping-redis` | 5초 | 3 | - | 수동 |
| `request-matcher` | 5분 | 3 | `15 * * * *` | `default` |
| `request-pruner` | 5분 | 3 | `0 0 * * *` | `default` |
| `runbox` | 5분 | 3 | - | `runbox`, API |
| `runbox-pruner` | 1분 | 3 | - | 수동 |
| `stat-{cf,ga,gsc,mw}-{daily,hourly}` | 5분 | 3 | `5 * * * *` | `default` |

`daily` 통계 task도 현재 매시 05분 실행되며 task 내부에서 수집 시간 범위를 결정한다.

### 운영 CLI

```bash
cd /app/goapp
go run ./cmd/ctl tasks
go run ./cmd/ctl tasks --watch
go run ./cmd/ctl ai-edit '{"task_id":123}'
go run ./cmd/ctl flush active
go run ./cmd/ctl flush pending
go run ./cmd/ctl flush scheduled
go run ./cmd/ctl flush retry
go run ./cmd/ctl flush all
```

`flush active`는 active task에 cancellation을 전달한다. `flush pending`, `flush scheduled`, `flush retry`는 해당 Asynq 상태의 task만 archive하며 `flush all`은 네 상태를 모두 처리한다. Asynq Redis key를 application code에서 직접 수정하지 말고 `asynq.Inspector` 또는 공식 도구를 사용한다.

## MediaWiki Extension 관리 CLI

`cmd/adm`은 GoApp task가 아니라 로컬 MediaWiki extension을 관리하는 관리자용 CLI다. `MW_INSTALL_PATH`가 MediaWiki 설치 경로를 가리켜야 하며, 이 값의 상위 디렉터리를 repository root로 사용한다. 일반 개발 환경에서는 `MW_INSTALL_PATH=/app/w`를 사용한다.

```bash
cd /app/goapp
go run ./cmd/adm extensions list
go run ./cmd/adm extensions upgrade
```

- `extensions list`: `w/extensions/*/extension.json`의 이름과 version을 읽고 `mwz/extensions/extensions.yaml`의 repo/tag 정보와 합쳐 출력한다. 설치 version과 지정 tag가 다르면 해당 행을 빨간색으로 표시한다.
- `extensions upgrade`: `mwz/extensions/extensions.yaml`을 기준으로 누락된 extension을 clone한다. 설치 version이 지정 tag와 다르면 기존 `w/extensions/<name>` 디렉터리를 삭제한 뒤 `git clone --depth=1 -b <tag>`로 다시 받는다.

`upgrade`는 관리 대상 extension 디렉터리의 로컬 변경도 함께 삭제하므로 실행 전에 필요한 변경이 없는지 확인한다. YAML에 없는 extension은 삭제하거나 갱신하지 않는다.
