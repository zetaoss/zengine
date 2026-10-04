# 설정

zengine 이미지는 설정을 두 가지로 받는다.

1. **환경변수**: goapp과 MediaWiki(PHP)가 프로세스 환경변수에서 읽는다. env 파일을 직접 읽지 않으므로 배포 환경이 컨테이너 환경변수로 넣는다. 예시는 루트의 `config.env.example`.
2. **외부에서 주입되는 파일**: 이 저장소에 없고, 배포 환경이 넣어 주는 파일. 어떻게 넣는지는 배포 환경이 정한다.

원칙

- 이 저장소는 파일마다 **자체 보유**인지 **외부 주입**(필수/선택)인지만 정한다.
- 비밀값은 이 저장소에 두지 않는다. 앱과 주입되는 PHP 설정은 비밀값을 **환경변수로 읽는다**. 파일로만 쓸 수 있는 비밀(GA 서비스 계정)은 그 **경로**를 환경변수로 받는다(`GA_READER_FILE`). 환경변수를 어떻게 채우는지는 배포 환경이 정한다.
- 환경마다 다른 값은 환경변수로 받는다. 환경별로 구조가 다른 설정만 주입 파일(`SiteSettings.php`)로 받는다.
- 사이트 고유의 운영 설정(확장 설정 등)은 공개하지 않는다. 주입 파일로 받는다.

## 환경변수

### goapp (`goapp/app/config/config.go`)

| 변수 | 기본값 | 설명 |
| --- | --- | --- |
| `APP_URL` | | 사이트 URL. MediaWiki `$wgServer`도 이 값을 쓴다 |
| `API_SERVER` | | 서버 쪽에서 MediaWiki API(`/w/api.php`)를 호출할 때 쓰는 내부 기준 URL (로그인 사용자 확인, 통계 수집 등) |
| `AVATAR_BASE_URL` | | 아바타 서비스 URL |
| `DEV_MODE` | `false` | 개발 모드(Vite dev 서버 프록시 등) |
| `INTERNAL_SECRET_KEY` | | 내부 API(`/api/internal/*`) 인증 키. 아바타 서비스와 공유 |
| `LOG_LEVEL` | `info` | 로그 레벨 |
| `DB_HOST`, `DB_PORT`(3306), `DB_DATABASE`, `DB_USERNAME`, `DB_PASSWORD` | | MariaDB |
| `REDIS_PERSIST_HOST`, `REDIS_PERSIST_PORT` | `REDIS_HOST`, `REDIS_PORT` | 퇴출되면 안 되는 데이터용 Redis: 작업 큐(Asynq), 인증 토큰, 요청 횟수 제한. [redis.md](redis.md) |
| `REDIS_VOLATILE_HOST`, `REDIS_VOLATILE_PORT` | `REDIS_HOST`, `REDIS_PORT` | 퇴출되어도 되는 데이터용 Redis: 캐시. [redis.md](redis.md) |
| `REDIS_HOST`, `REDIS_PORT` | (6379) | 이전의 단일 Redis 설정. 역할별 변수가 없을 때만 쓰인다 |
| `AD_CLIENT`, `AD_SLOTS` | | 광고. `AD_SLOTS`는 쉼표 구분 |
| `GA_MEASUREMENT_ID`, `GA_PROPERTY_ID`, `GA_TIMEZONE`, `GSC_SITE_URL` | | Google Analytics / Search Console |
| `GA_READER_FILE` | | GA/GSC 조회용 서비스 계정 JSON 파일 경로 |
| `LLM_ENDPOINT`, `RUNBOX_ENDPOINT`, `SEARCH_ENDPOINT` | | 외부 API |
| `MONITORING_ENDPOINT`, `MONITORING_NAMESPACE`, `MONITORING_NODEPOOL`, `MONITORING_PVC` | | 모니터링 조회 |
| `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_ZONE_ID` | | Cloudflare API |
| `FACEBOOK_CLIENT_ID/SECRET`, `GITHUB_CLIENT_ID/SECRET`, `GOOGLE_CLIENT_ID/SECRET` | | 소셜 로그인 |

goapp은 시작할 때 이 값을 읽어 프런트엔드에 `window.ZCONF`(`avatarBaseUrl`, `gaMeasurementId`, `adClient`, `adSlots`)로 넣는다(`goapp/server/runtime/common/injector.go`).

### MediaWiki (PHP)

대부분 `BaseSettings.php`가 읽는다. 기본값이 없으므로 모두 설정한다(`MW_CDN_SERVERS`, `AD_*` 등 목록·선택 값은 비워도 된다). `BaseSettings.php`는 goapp과 달리 `REDIS_HOST`/`REDIS_PORT`로 대체하지 않는다.

| 변수 | 사용처 |
| --- | --- |
| `APP_URL` | `$wgServer` |
| `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD` | `$wgDBserver`, `$wgDBuser`, `$wgDBpassword`. DB 이름은 `zetawiki`로 고정(goapp도 `zetawiki.*`로 참조) |
| `REDIS_VOLATILE_HOST`, `REDIS_VOLATILE_PORT` | 캐시 `$wgObjectCaches['redis-cache']`. [redis.md](redis.md) |
| `REDIS_PERSIST_HOST`, `REDIS_PERSIST_PORT` | 세션 `$wgObjectCaches['redis-session']`, 작업 큐 `$wgJobTypeConf`. ZetaExtension 인증 상태(OTP, 소셜 로그인 연계, `includes/Auth/PersistRedis.php`)도 여기서 goapp이 쓴 토큰을 읽는다 |
| `MW_SECRET_KEY`, `MW_UPGRADE_KEY` | `$wgSecretKey`, `$wgUpgradeKey` (비밀) |
| `MW_CDN_SERVERS` | `$wgCdnServers` (쉼표 구분) |
| `SHELLBOX_SCORE_URL`, `SHELLBOX_SECRET_KEY` | `$wgShellboxUrls['score']`, `$wgShellboxSecretKey` (키는 비밀) |
| `AVATAR_BASE_URL`, `GA_MEASUREMENT_ID`, `AD_CLIENT`, `AD_SLOTS` | 스킨 상수(아래) |
| `MW_INSTALL_PATH` | ZetaExtension 유지보수 스크립트. 이미지가 MediaWiki 디렉터리로 설정한다(`dev`: `/var/www/html`, `prod`: `/app/w`) |

## MediaWiki 설정 파일

모두 MediaWiki 디렉터리(`$IP`)에 놓인다. `LocalSettings.php`가 진입점이며 아래 순서로 `require`한다. 뒤에 오는 파일이 앞의 값을 덮어쓴다.

| 순서 | 파일 | 구분 | 역할 |
| --- | --- | --- | --- |
| - | `LocalSettings.php` | 자체 보유 (`mwz/settings/`) | 진입점. 아래 파일을 순서대로 부른다 |
| 1 | `BaseSettings.php` | 자체 보유 (`mwz/settings/`) | 모든 환경 공통: 사이트, DB, 캐시·작업 큐, CDN, 업로드, 권한, 스킨(ZetaSkin)과 스킨 상수. 값과 비밀값은 환경변수(위) |
| 2 | `SiteSettings.php` | 외부 주입(필수) | 환경별 설정(확장 제외). 예: 개발 환경의 디버그 설정, `$wgCdnServersNoPurge` |
| 3 | `ExtensionSettings.php` | 외부 주입(필수) | 모든 확장(기본 포함, 외부, ZetaExtension)의 `wfLoadExtension`과 설정 |

이미지(`base` 단계)는 `LocalSettings.php`와 `BaseSettings.php`를 MediaWiki 디렉터리에 넣는다. 배포 환경은 `SiteSettings.php`와 `ExtensionSettings.php`를 같은 디렉터리에 넣는다. 둘 중 하나라도 없으면 MediaWiki가 시작하지 않는다. 내용이 없으면 `<?php`만 있는 파일을 둔다.

### 스킨 상수

ZetaSkin(`mwz/skins/ZetaSkin/includes/SkinZetaSkin.php`)이 쓰는 PHP 상수는 `BaseSettings.php`가 정의한다.

| 상수 | 값 |
| --- | --- |
| `ASSET_HASH` | 스킨 번들(`skins/ZetaSkin/dist/app.js`)의 수정 시각. 번들을 다시 빌드하면 바뀐다. 파일이 없으면 현재 시각 |
| `AVATAR_BASE_URL`, `GA_MEASUREMENT_ID`, `AD_CLIENT` | 같은 이름의 환경변수 |
| `AD_SLOTS` | 환경변수 `AD_SLOTS`(쉼표 구분)를 JSON 배열로 바꾼 값 |

## MediaWiki 확장

- **외부 확장**(git에서 받는, MediaWiki 기본 포함이 아닌 확장): `mwz/extensions.yaml`이 설치 목록이다. 목록에 있으면 이미지(`base` 단계)에 설치된다. 빼려면 주석 처리한다. 확장마다 출처(`repo`/`tag`)를 둔다. 설치되는 commit은 `mwz/extensions.lock`에 고정되고(`make extensions-lock`), 브랜치(`REL1_43` 등)가 움직여도 lock을 갱신하기 전까지는 같은 commit이 설치된다. PHP 의존성은 `hack/mediawiki-composer.lock`.
- **ZetaExtension**: 이 저장소의 일부(`mwz/extensions/ZetaExtension`)라 설치 목록에 없다. **ZetaSkin**은 스킨이라 `BaseSettings.php`가 로드한다.
- **기본 포함 확장**(Cite, VisualEditor 등): 설치할 것이 없다.
- **로드와 설정**: 모든 확장의 `wfLoadExtension`과 설정(`$wg…`)은 주입되는 `ExtensionSettings.php`가 맡는다. 이 저장소는 설치만 한다. 설치하고 로드하지 않은 확장은 쓰이지 않을 뿐이다. 확장 패키지와 로드·설정을 느슨하게 묶어, 사이트 고유의 설정을 공개하지 않는다.

## 그 밖의 외부 주입 파일

| 파일 | 목표 | 현재 | 내용 |
| --- | --- | --- | --- |
| 컨테이너 시작 스크립트 | 자체 보유 (dev/prod별) | 외부 주입 | 설정 파일 배치, 서비스 시작 |
| `nginx.conf`, `php-fpm.conf`, `php.ini` | 자체 보유 (dev/prod별) | 외부 주입 | 웹 서버, PHP |
| `supervisord.conf` | 자체 보유 (dev) | 외부 주입 | 개발 이미지의 프로세스 구성 |
| `dist_ads.txt`, `dist_robots.txt` | 자체 보유 | 외부 주입 | 정적 파일(`/app/svelte/dist/`) |
| `dist_config.js` | 없앰 | 외부 주입 | 쓰이지 않는 것으로 보임. goapp이 같은 값을 `window.ZCONF`로 넣는다 |
| `SyntaxHighlight.php`, `MsUpload.less`, `mediawiki.skin.defaults.less` | 자체 보유 | 외부 주입 | MediaWiki 패치 |
| GA 서비스 계정 JSON | 외부 주입(선택) | 외부 주입 | 비밀 파일. `GA_READER_FILE`이 경로를 가리킨다 |

목표 구조로 옮기면 외부 주입 파일은 `SiteSettings.php`, `ExtensionSettings.php`, 비밀 파일(GA 서비스 계정)만 남는다.
