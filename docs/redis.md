# Redis

zengine uses Redis for two roles, based on whether losing or evicting the data would cause a problem.

| Role | Requirements | Environment variable | Data |
| --- | --- | --- | --- |
| persist | Must not be evicted or lost. Use a `noeviction` policy and persistence (such as AOF). | `REDIS_PERSIST_HOST` | goapp task queues (Asynq: server, worker, scheduler, tool), OTP and social-login tokens, rate-limit counters, MediaWiki sessions, MediaWiki job queue |
| volatile | Data may be evicted at any time. Policies such as `allkeys-lru` are suitable; persistence is unnecessary. | `REDIS_VOLATILE_HOST` | goapp MediaWiki user cache (1-minute TTL), MediaWiki caches (main, message, language converter) |

The MediaWiki parser cache is stored in MariaDB's `objectcache` table with a 30-day TTL, rather than in either Redis role. Because this cache is disposable, deployment environments should consider excluding its rows from database backups.

- Persist data has a TTL or is deleted after processing, so it does not grow indefinitely. If `noeviction` Redis runs out of memory, writes fail visibly instead of silently losing data.
- If persist data were operated like a cache (for example, with `allkeys-lru`), queued jobs could disappear before execution and login tokens could disappear, causing login failures.
- Both Redis roles can use the same Redis instance. In that case, follow the persist requirements and use `noeviction`.
- Choose a role for each new use based on these requirements.

## Environment variables

| Variable | Default |
| --- | --- |
| `REDIS_PERSIST_HOST` | goapp uses `127.0.0.1`; MediaWiki requires a host to be configured |
| `REDIS_VOLATILE_HOST` | goapp uses `127.0.0.1`; MediaWiki requires a host to be configured |

Both Redis roles use port `6379`. Set the host variables to a hostname or IP address only; `host:port` and Redis URIs are not supported.

## Code

| Use | Role | Location |
| --- | --- | --- |
| Asynq task queue | persist | `goapp/app/redis` `AsynqConnOpt` (server, worker, scheduler, tool) |
| Social-login tokens (write) | persist | `goapp/server/handlers/auth/social` → `OpenPersist` |
| Rate limiting | persist | `goapp/server/throttle` → `OpenPersist` |
| MediaWiki user cache | volatile | `goapp/server/auth` → `OpenVolatile` |
| OTP and linked-login tokens (read) | persist | ZetaExtension `includes/Auth/PersistRedis.php` |
| `ping-redis` health-check task | both | `goapp/tasks/pingredis` |

## MediaWiki

MediaWiki reads its Redis connection settings from the environment variables above in `mwz/settings/BaseSettings.php` ([config.md](config.md)). The roles are:

| MediaWiki setting | Role |
| --- | --- |
| `$wgObjectCaches['redis-volatile']` (main, message, and language-converter caches) | volatile |
| `$wgParserCacheType` (`CACHE_DB`, `$wgParserCacheExpireTime = 86400 * 30`) | MariaDB `objectcache` |
| `$wgObjectCaches['redis-persist']` (`$wgSessionCacheType`) | persist |
| `$wgJobTypeConf['default']` (`JobQueueRedis`) | persist |

`redis-volatile` and `redis-persist` are cache configuration IDs in MediaWiki; each reads its host from the corresponding environment variable.
