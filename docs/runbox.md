# Runbox

문서의 실행 가능한 코드 블록을 외부 Runbox 서버에서 실행하고 결과를 페이지에 표시하는 기능이다.

## 흐름

```text
MediaWiki 코드 블록
  -> Skin Svelte
  -> Go API (/api/runbox)
  -> runboxes 테이블 / Asynq runbox queue
  -> worker
  -> RUNBOX_ENDPOINT/{lang|notebook}
  -> 결과 polling 및 화면 표시
```

`run` 속성이 있는 일반 코드 블록과 `notebook` 속성이 있는 노트북 블록을 지원한다. `javascript`, `html`, `css`는 브라우저에서 직접 렌더링하고, 그 외 언어는 외부 Runbox 서버로 보낸다.

## API

라우트는 `goapp/server/routes.go`에 정의되어 있다.

| Method | Path | 설명 |
| --- | --- | --- |
| `GET` | `/api/runbox/{hash}` | 실행 상태와 결과 조회 |
| `POST` | `/api/runbox` | 실행 요청 생성 |
| `POST` | `/api/runbox/{hash}/rerun` | sysop용 재실행 |

실행 상태는 `pending`, `running`, `succeeded`, `failed` 중 하나다. 실패 시 외부 서버의 오류 사유를 `outs.error`에 저장하고 화면에 표시한다.

일반 실행 payload 예시:

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

노트북 payload는 `lang`과 코드 셀 순서의 `sources` 배열을 사용한다.

## Worker와 설정

worker는 전용 `runbox` queue에서 task를 처리하며 retry하지 않는다. 외부 요청 timeout은 60초, task timeout은 2분이다. 실행을 시작한 뒤 오래 갱신되지 않은 `running` 작업은 pruner가 3분 후 실패 처리한다. queue에서 대기 중인 `pending` 작업은 정상적인 지연일 수 있으므로 pruner가 임의로 실패 처리하지 않는다.

`.env` 또는 `ENV_FILE`에 외부 서버 주소를 설정한다.

```dotenv
RUNBOX_ENDPOINT=https://runbox.example.internal
```

endpoint 끝에는 `/`를 붙이지 않는다. worker는 다음 경로로 JSON POST 요청을 보낸다.

```text
RUNBOX_ENDPOINT/lang
RUNBOX_ENDPOINT/notebook
```

외부 서버는 2xx 응답과 JSON object를 반환해야 한다. 일반 실행 결과는 `logs`, `images`, 노트북 결과는 `outputsList`를 사용한다.

## 관련 코드

- API: `goapp/server/handlers/api/runbox/runbox.go`
- task와 stale 작업 정리: `goapp/tasks/runbox/`
- task 등록: `goapp/worker/registry/registry.go`
- 설정: `goapp/app/config/config.go`, `.env.example`
- 프론트: `mwz/skins/ZetaSkin/svelte/src/components/runbox/`
