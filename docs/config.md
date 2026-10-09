# Configuration

The zengine image receives configuration in two forms:

1. **Environment variables**: goapp and MediaWiki (PHP) read values from the process environment. They do not read an env file directly; the deployment environment provides container variables. See `config.env.example` in the repository root.
2. **Externally mounted files**: Files that are not stored in this repository and are provided by the deployment environment. Each deployment decides how to provide them.

## Principles

- This repository specifies whether each file is **maintained here** or **provided externally** (required or optional).
- Do not store secrets in this repository. The application and mounted PHP settings read secrets from **environment variables**. The deployment environment supplies the variable values. Credentials for external analytics APIs (Cloudflare, Google) live in bob, not here.
- Use environment variables for values that differ between environments. Use mounted files (`SiteSettings.php`) only when the configuration structure differs by environment.
- Do not expose site-specific operational settings, such as extension configuration. Provide them through mounted files.

## Environment variables

### goapp (`goapp/app/config/config.go`)

| Variable | Default | Description |
| --- | --- | --- |
| `APP_URL` | | Site URL. MediaWiki also uses this for `$wgServer`. |
| `API_SERVER` | | Internal base URL used by the server to call the MediaWiki API (`/w/api.php`), for example to check the logged-in user or collect statistics. |
| `AVATAR_BASE_URL` | | Avatar service URL. |
| `DEV_MODE` | `false` | Development mode, including the Vite dev-server proxy. |
| `INTERNAL_SECRET_KEY` | | Authentication key for internal APIs (`/api/internal/*`), shared with the avatar service. |
| `LOG_LEVEL` | `info` | Log level. |
| `DB_HOST`, `DB_PORT` (`3306`), `DB_DATABASE`, `DB_USERNAME`, `DB_PASSWORD` | | MariaDB connection settings. |
| `REDIS_PERSIST_HOST` | goapp defaults to `127.0.0.1`; MediaWiki requires a host to be configured (port `6379` is fixed) | Redis for data that must not be evicted: task queues (Asynq), authentication tokens, and rate limits. See [redis.md](redis.md). |
| `REDIS_VOLATILE_HOST` | goapp defaults to `127.0.0.1`; MediaWiki requires a host to be configured (port `6379` is fixed) | Redis for data that may be evicted: caches. See [redis.md](redis.md). |
| `AD_CLIENT`, `AD_SLOTS` | | Advertising. `AD_SLOTS` is comma-separated. |
| `GA_MEASUREMENT_ID` | | Google Analytics measurement ID for the frontend tag. |
| `BOB_ENDPOINT` | | Base URL of [bob](https://github.com/zetaoss/bob), the in-cluster app server, without a trailing `/`. goapp appends the feature path: `/aigate` (LLM), `/cloudflare/analytics` (`stat-cf-*`), `/ga/report` (`stat-ga-*`), `/gsc/query` (`stat-gsc-*`), `/metrics/` (k8s stats for `stat-k8s-hourly` and `ctl metrics`), `/runbox`, `/search`. |
| `FACEBOOK_CLIENT_ID/SECRET`, `GITHUB_CLIENT_ID/SECRET`, `GOOGLE_CLIENT_ID/SECRET` | | Social login. |

At startup, goapp reads these values and injects `window.ZCONF` (`avatarBaseUrl`, `gaMeasurementId`, `adClient`, `adSlots`) into the frontend (`goapp/server/runtime/common/injector.go`).

### MediaWiki (PHP)

Most variables are read by `BaseSettings.php`. Set all variables without defaults; list and optional values such as `MW_CDN_SERVERS` and `AD_*` may be empty. goapp and MediaWiki both use the role-specific Redis variables.

| Variable | Use |
| --- | --- |
| `APP_URL` | `$wgServer` |
| `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD` | `$wgDBserver`, `$wgDBuser`, `$wgDBpassword`. The database name is fixed as `zetawiki` (goapp also refers to tables as `zetawiki.*`). |
| `REDIS_VOLATILE_HOST` | Cache `$wgObjectCaches['redis-volatile']`. Port `6379` is fixed. See [redis.md](redis.md). |
| `REDIS_PERSIST_HOST` | Session cache `$wgObjectCaches['redis-persist']` and job queue `$wgJobTypeConf`. ZetaExtension authentication state (OTP and linked social logins, `includes/Auth/PersistRedis.php`) also reads tokens written by goapp from here. Port `6379` is fixed. |
| `MW_SECRET_KEY`, `MW_UPGRADE_KEY` | `$wgSecretKey`, `$wgUpgradeKey` (secrets). |
| `MW_CDN_SERVERS` | `$wgCdnServers` (comma-separated). |
| `SHELLBOX_SCORE_URL`, `SHELLBOX_SECRET_KEY` | `$wgShellboxUrls['score']`, `$wgShellboxSecretKey` (the key is secret). |
| `AVATAR_BASE_URL`, `GA_MEASUREMENT_ID`, `AD_CLIENT`, `AD_SLOTS` | Skin constants (see below). |
| `MW_INSTALL_PATH` | ZetaExtension maintenance scripts. The image sets this to the MediaWiki directory (`dev`: `/var/www/html`, `prod`: `/app/w`). |

## MediaWiki settings files

All files are placed in the MediaWiki directory (`$IP`). `LocalSettings.php` is the entry point and requires the following files in order. Later files override earlier values.

| Order | File | Source | Purpose |
| --- | --- | --- | --- |
| - | `LocalSettings.php` | Maintained here (`mwz/settings/`) | Entry point; loads the files below in order. |
| 1 | `BaseSettings.php` | Maintained here (`mwz/settings/`) | Shared site, database, cache and job-queue, CDN, upload, permission, and skin settings, including skin constants. Values and secrets come from environment variables (above). |
| 2 | `SiteSettings.php` | Externally provided (required) | Environment-specific settings, excluding extension configuration. Examples include development debug settings and `$wgCdnServersNoPurge`. |
| 3 | `ExtensionSettings.php` | Externally provided (required) | `wfLoadExtension` calls and settings for all extensions (bundled, external, and ZetaExtension). |

The image's `base` stage places `LocalSettings.php` and `BaseSettings.php` in the MediaWiki directory. The deployment environment places `SiteSettings.php` and `ExtensionSettings.php` in the same directory. MediaWiki will not start if either file is missing. If a file has no settings, provide a file containing only `<?php`.

### Skin constants

`BaseSettings.php` defines the PHP constants used by ZetaSkin (`mwz/skins/ZetaSkin/includes/SkinZetaSkin.php`).

| Constant | Value |
| --- | --- |
| `ASSET_HASH` | Modification time of the skin bundle (`skins/ZetaSkin/dist/app.js`); it changes when the bundle is rebuilt. Uses the current time if the file is missing. |
| `AVATAR_BASE_URL`, `GA_MEASUREMENT_ID`, `AD_CLIENT` | Environment variables with the same names. |
| `AD_SLOTS` | The `AD_SLOTS` environment variable, converted from comma-separated values to a JSON array. |

## MediaWiki extensions

- **External extensions** (not bundled with MediaWiki, fetched from Git): `mwz/extensions.yaml` lists what is installed. Listed extensions are installed in the image's `base` stage; comment out an entry to remove it. Each extension specifies a source (`repo`/`tag`). The installed commit is pinned in `mwz/extensions.lock` (`make extensions-lock`), so a moving branch such as `REL1_43` stays on the same commit until the lock is updated. PHP dependencies are pinned in `hack/mediawiki-composer.lock`.
- **ZetaExtension**: Part of this repository (`mwz/extensions/ZetaExtension`), so it is not listed for external installation. **ZetaSkin** is a skin and is loaded by `BaseSettings.php`.
- **Bundled extensions** (Cite, VisualEditor, etc.): Nothing needs to be installed.
- **Loading and configuration**: The externally provided `ExtensionSettings.php` contains every extension's `wfLoadExtension` call and `$wg…` settings. This repository only installs the extensions. An installed extension that is not loaded is simply unused. Keeping installation separate from loading and configuration avoids exposing site-specific settings.

## Other externally provided files

MediaWiki patches are included in the image ([patches.md](patches.md)).

| File | Target source | Current source | Purpose |
| --- | --- | --- | --- |
| Container startup script | Maintained here (dev/prod-specific) | Provided externally | Place configuration files and start services. |
| `nginx.conf`, `php-fpm.conf`, `php.ini` | Maintained here (dev/prod-specific) | Provided externally | Web server and PHP configuration. |
| `supervisord.conf` | Maintained here (dev) | Provided externally | Process configuration for the development image. |
| `dist_ads.txt`, `dist_robots.txt` | Maintained here | Provided externally | Static files (`/app/svelte/dist/`). |

After moving to the target structure, the only externally provided files will be `SiteSettings.php` and `ExtensionSettings.php`.
