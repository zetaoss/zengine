# Redis

zengine은 Redis를 데이터의 성격에 따라 두 역할로 나눠 쓴다. 기준은 "퇴출되거나 사라지면 문제가 되는가"이다.

| 역할 | 요구 | 환경변수 | 데이터 |
| --- | --- | --- | --- |
| persist | 퇴출·유실되면 안 된다. 퇴출 정책은 `noeviction`, 영속화(AOF 등) 권장 | `REDIS_PERSIST_HOST`, `REDIS_PERSIST_PORT` | goapp 작업 큐(Asynq: server, worker, scheduler, tool), OTP·소셜 로그인 연계 토큰, 요청 횟수 제한 카운터, MediaWiki 세션, MediaWiki 작업 큐 |
| volatile | 언제든 퇴출되어도 된다. 퇴출 정책은 `allkeys-lru` 등, 영속화 불필요 | `REDIS_VOLATILE_HOST`, `REDIS_VOLATILE_PORT` | goapp의 MediaWiki 사용자 캐시(TTL 1분), MediaWiki 캐시(main, message, parser, language converter) |

- persist의 데이터는 TTL이 있거나 처리 후 지워지므로 계속 쌓이지 않는다. `noeviction`에서 메모리가 차면 쓰기가 오류로 실패하므로, 조용히 데이터를 잃는 대신 바로 드러난다.
- persist를 캐시처럼(`allkeys-lru` 등) 운영하면 대기 중인 작업이 실행 전에 사라지고, 로그인 연계 토큰이 사라져 로그인이 실패할 수 있다.
- 두 역할을 같은 Redis 하나로 운영해도 된다. 이때는 persist의 요구(`noeviction`)를 따른다.
- 새 용도를 붙일 때는 위 기준으로 역할을 고른다.

## 환경변수

| 변수 | 기본값 |
| --- | --- |
| `REDIS_PERSIST_HOST`, `REDIS_PERSIST_PORT` | 없으면 `REDIS_HOST`, `REDIS_PORT` |
| `REDIS_VOLATILE_HOST`, `REDIS_VOLATILE_PORT` | 없으면 `REDIS_HOST`, `REDIS_PORT` |
| `REDIS_HOST`, `REDIS_PORT`(6379) | 이전의 단일 Redis 설정. 역할별 변수가 없을 때만 쓰인다. 모든 배포가 역할별 변수로 옮기면 제거한다 |

`REDIS_*_HOST`에는 호스트 이름 대신 `host:port`나 `redis://…` URI를 줄 수도 있다(goapp).

## 코드

| 사용처 | 역할 | 위치 |
| --- | --- | --- |
| Asynq 작업 큐 | persist | `goapp/app/redis` `AsynqConnOpt` (server, worker, scheduler, tool) |
| 소셜 로그인 연계 토큰 (쓰기) | persist | `goapp/server/handlers/auth/social` → `OpenPersist` |
| 요청 횟수 제한 | persist | `goapp/server/throttle` → `OpenPersist` |
| MediaWiki 사용자 캐시 | volatile | `goapp/server/auth` → `OpenVolatile` |
| OTP·연계 토큰 (읽기) | persist | ZetaExtension `includes/Auth/PersistRedis.php` |
| 상태 확인 작업 `ping-redis` | 둘 다 | `goapp/tasks/pingredis` |

## MediaWiki

MediaWiki의 Redis 연결은 `mwz/settings/BaseSettings.php`가 위 환경변수에서 읽는다([config.md](config.md)). 역할은 다음과 같다.

| MediaWiki 설정 | 역할 |
| --- | --- |
| `$wgObjectCaches['redis-cache']` (main, message, parser, language converter 캐시) | volatile |
| `$wgObjectCaches['redis-session']` (`$wgSessionCacheType`) | persist |
| `$wgJobTypeConf['default']` (`JobQueueRedis`) | persist |

`redis-cache`, `redis-session`은 MediaWiki 안의 캐시 이름일 뿐 Redis 서버 이름과는 관계없다.
