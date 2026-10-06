# Runbox

This feature runs executable code blocks from wiki pages on an external Runbox server and displays the results on the page.

## Flow

```text
MediaWiki code block
  -> Skin Svelte
  -> Go API (/api/runbox)
  -> runboxes table / Asynq runbox queue
  -> worker
  -> RUNBOX_ENDPOINT/{lang|notebook}
  -> result polling and display
```

The feature supports regular code blocks with the `run` attribute and notebook blocks with the `notebook` attribute. `javascript`, `html`, and `css` run directly in the browser; other languages are sent to the external Runbox server.

## API

Routes are defined in `goapp/server/routes.go`.

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/runbox/{hash}` | Get execution status and results |
| `POST` | `/api/runbox` | Create an execution request |
| `POST` | `/api/runbox/{hash}/rerun` | Rerun as a sysop |

Execution status is one of `pending`, `running`, `succeeded`, or `failed`. On failure, the external server's error is stored in `outs.error` and displayed on the page.

Example payload for a regular execution:

```json
{
  "hash": "SHA-256 hex string",
  "page_id": 123,
  "type": "lang",
  "payload": {
    "lang": "python",
    "files": [{"name": "main.py", "body": "print('hello')"}],
    "main": 0
  }
}
```

A notebook payload uses `lang` and a `sources` array ordered by code cell.

## Worker and configuration

The worker processes tasks on the dedicated `runbox` queue and does not retry them. The external request timeout is 60 seconds, and the task timeout is 2 minutes. A `running` task that has not been updated for 3 minutes after starting is marked as failed by the pruner. A `pending` task waiting in the queue may be delayed normally, so the pruner does not fail it automatically.

Set the external server address in an environment variable:

```dotenv
RUNBOX_ENDPOINT=https://runbox.example.internal
```

Do not add a trailing `/` to the endpoint. The worker sends JSON POST requests to these paths:

```text
RUNBOX_ENDPOINT/lang
RUNBOX_ENDPOINT/notebook
```

The external server must return a 2xx response and a JSON object. Regular execution results use `logs` and `images`; notebook results use `outputsList`.

## Related code

- API: `goapp/server/handlers/api/runbox/runbox.go`
- Tasks and stale-task pruning: `goapp/tasks/runbox/`
- Task registration: `goapp/worker/registry/registry.go`
- Configuration: `goapp/app/config/config.go`, `config.env.example`
- Frontend: `mwz/skins/ZetaSkin/svelte/src/components/runbox/`
