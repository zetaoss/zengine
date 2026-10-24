# Configuration

The zengine image receives configuration in three forms:

1. **Environment variables**: goapp and MediaWiki (PHP) read values from the process environment. They do not read an env file directly; the deployment environment provides container variables. See `config.env.example` in the repository root.
2. **goapp config file**: an optional YAML file named by `GOAPP_CONFIG` for goapp settings that are neither secrets nor endpoints (see [goapp config file](#goapp-config-file)).
3. **Externally mounted files**: Files that are not stored in this repository and are provided by the deployment environment. Each deployment decides how to provide them.

## Principles

- This repository specifies whether each file is **maintained here** or **provided externally** (required or optional).
- Do not store secrets in this repository. The application and mounted PHP settings read secrets from **environment variables**. The deployment environment supplies the variable values. Credentials for external analytics APIs (Cloudflare, Google) live in bob, not here.
- Use environment variables for values that differ between environments. Use mounted files (`SiteSettings.php`) only when the configuration structure differs by environment.
- Extension loading and general extension settings are maintained here (`BaseExtensionSettings.php`), next to the installation list, so adding or removing an extension is one change and can be tried with this repository alone. Site-specific parts are loaded and configured by the mounted `SiteExtensionSettings.php`, so they are not exposed: abuse controls (AbuseFilter, auto-confirm thresholds) and extensions that depend on site services (S3 file storage, mail, Score's Shellbox).

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
| `DB_HOST`, `DB_DATABASE`, `DB_USERNAME`, `DB_PASSWORD` | | MariaDB connection settings. The port is fixed at `3306`. |
| `REDIS_PERSIST_HOST` | goapp defaults to `127.0.0.1`; MediaWiki requires a host to be configured (port `6379` is fixed) | Redis for data that must not be evicted: task queues (Asynq), authentication tokens, and rate limits. See [redis.md](redis.md). |
| `REDIS_VOLATILE_HOST` | goapp defaults to `127.0.0.1`; MediaWiki requires a host to be configured (port `6379` is fixed) | Redis for data that may be evicted: caches. See [redis.md](redis.md). |
| `AD_CLIENT`, `AD_SLOTS` | | Advertising. `AD_SLOTS` is comma-separated. |
| `GA_MEASUREMENT_ID` | | Google Analytics measurement ID for the frontend tag. |
| `BOB_ENDPOINT` | | Base URL of [bob](https://github.com/zetaoss/bob), the in-cluster app server, without a trailing `/`. goapp appends the feature path: `/aigate` (LLM), `/cloudflare/analytics` (`stat-cf-*`), `/ga/report` (`stat-ga-*`), `/gsc/query` (`stat-gsc-*`), `/metrics/` (k8s stats for `stat-k8s-hourly` and `ctl metrics`), `/runbox`, `/search`. |
| `FACEBOOK_CLIENT_ID/SECRET`, `GITHUB_CLIENT_ID/SECRET`, `GOOGLE_CLIENT_ID/SECRET` | | Social login. |
| `GOAPP_CONFIG` | | Path of the [goapp config file](#goapp-config-file). Unset means no file; a set path that cannot be read or parsed stops goapp at startup. |

At startup, goapp reads these values and injects `window.ZCONF` (`avatarBaseUrl`, `gaMeasurementId`, `adClient`, `adSlots`) into the frontend (`goapp/server/runtime/common/injector.go`).

### goapp config file

A YAML file (`goapp/app/config/config.go`, `GOAPP_CONFIG`) for goapp settings that are neither secrets nor endpoints, so they do not each need an environment variable. Unknown keys are errors, so a typo does not silently keep a default. Sections:

- `chains`: run settings of the chain loops, `common` plus `docDefine` and `docEval` overriding it field by field. See [doc-eval.md](doc-eval.md#settings).

```yaml
chains:
  common:
    model: groq/openai/gpt-oss-120b
    strategy: random
  docDefine:
    interval: 30m
  docEval:
    interval: 1h
```

### MediaWiki (PHP)

Most variables are read by `BaseSettings.php`. Set all variables without defaults; list and optional values such as `MW_CDN_SERVERS` and `AD_*` may be empty. goapp and MediaWiki both use the role-specific Redis variables.

| Variable | Use |
| --- | --- |
| `APP_URL` | `$wgServer` |
| `DB_HOST`, `DB_USERNAME`, `DB_PASSWORD` | `$wgDBserver`, `$wgDBuser`, `$wgDBpassword`. The port is fixed as `3306` and the database name as `zetawiki` (goapp also refers to tables as `zetawiki.*`). |
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
| 3 | `BaseExtensionSettings.php` | Maintained here (`mwz/settings/`) | `wfLoadExtension` calls for the bundled and external extensions and ZetaExtension, except the site-specific ones below. |
| 4 | `SiteExtensionSettings.php` | Externally provided (required) | Site-specific extensions, loaded and configured together: AbuseFilter with the auto-confirm thresholds, AWS (S3 file repository), MailAPI and Score. |

The image places `LocalSettings.php`, `BaseSettings.php` and `BaseExtensionSettings.php` in the MediaWiki directory (`/app/w`), after moving MediaWiki there, so a settings change does not rebuild the MediaWiki layer. The `dev` image does not include them; the development environment links them from the working tree. The deployment environment places `SiteSettings.php` and `SiteExtensionSettings.php` in the same directory. MediaWiki will not start if either file is missing. If a file has no settings, provide a file containing only `<?php`.

### Notes on `BaseSettings.php`

PHP settings files carry no comments; the reasons behind non-obvious values are kept here.

- **Localisation cache**: `$wgLocalisationCacheConf['store'] = 'files'` with `$wgCacheDirectory = "$IP/cache"` keeps the cache in container-local CDB files instead of the `l10n_cache` table. Each container fills it on first use and rebuilds it when message files change. Not `'array'`: OPcache does not revalidate timestamps in production, so rebuilt PHP files would not be picked up.
- **`$wgAuthenticationTokenVersion`**: mixed into every session and "keep me logged in" token. Bumping it (for example `'1'` → `'2'`) logs every user out, e.g. after a session or `$wgSecretKey` leak. Removing it would also log everyone out once, since the default is `null`.

### Skin constants

`BaseSettings.php` defines the PHP constants used by ZetaSkin (`mwz/skins/ZetaSkin/includes/SkinZetaSkin.php`).

| Constant | Value |
| --- | --- |
| `AVATAR_BASE_URL`, `GA_MEASUREMENT_ID`, `AD_CLIENT` | Environment variables with the same names. |
| `AD_SLOTS` | The `AD_SLOTS` environment variable, split on commas into an array. |

The skin passes these to the frontend as `window.ZCONF`, built with `json_encode` so the values are escaped.

The skin's bundle is built with content-hashed names (`dist/app.<hash>.js`, `dist/app.<hash>.css`), and the build writes `dist/assets.php` with those names and the content hash of the main app's `/track.js`. `SkinZetaSkin` requires that file (cached by OPcache), so a changed file always gets a new URL and the files can be cached as immutable. The Vite dev server writes `assets.php` too, pointing at the dev stubs `app.js`/`app.css` with their content hash as the query string; without either a build or the dev server the skin fails to render.

## MediaWiki extensions

- **External extensions** (not bundled with MediaWiki, fetched from Git): `mwz/extensions.yaml` lists what is installed. Listed extensions are installed in the image's `base` stage; comment out an entry to remove it. Each extension specifies a source (`repo`/`tag`). The installed commit is pinned in `mwz/extensions.lock` (`make extensions-lock`), so a moving branch such as `REL1_43` stays on the same commit until the lock is updated. PHP dependencies are pinned in `hack/mediawiki-composer.lock`.
- **ZetaExtension**: Part of this repository (`mwz/extensions/ZetaExtension`), so it is not listed for external installation. **ZetaSkin** is a skin and is loaded by `BaseSettings.php`.
- **Bundled extensions** (Cite, VisualEditor, etc.): Nothing needs to be installed.
- **Loading and configuration**: `mwz/settings/BaseExtensionSettings.php` loads the extensions; adding an external extension means an entry in `mwz/extensions.yaml` and a `wfLoadExtension` line there. AbuseFilter, AWS, MailAPI and Score are site-specific, so the externally provided `SiteExtensionSettings.php` loads and configures them instead; without it they are not loaded. An installed extension that is not loaded is simply unused.
- **Math** is not loaded; SimpleMathJax renders `<math>` instead.

## Other externally provided files

MediaWiki patches are included in the image ([patches.md](patches.md)).

| File | Target source | Current source | Purpose |
| --- | --- | --- | --- |
| Container startup script | Maintained here (dev/prod-specific) | Provided externally | Place configuration files and start services. |
| `nginx.conf`, `php-fpm.conf`, `php.ini` | Maintained here (dev/prod-specific) | Provided externally | Web server and PHP configuration. |
| `supervisord.conf` | Maintained here (dev) | Provided externally | Process configuration for the development image. |
| `dist_ads.txt`, `dist_robots.txt` | Maintained here | Provided externally | Static files (`/app/svelte/dist/`). |

After moving to the target structure, the only externally provided files will be `SiteSettings.php` and `SiteExtensionSettings.php`.
