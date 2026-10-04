# 설정

zengine 이미지는 설정을 두 경로로 받는다.

1. **환경변수**: goapp과 MediaWiki(PHP)가 읽는다. 예시는 루트의 `.env.example`.
2. **배포 환경이 제공하는 파일**: 컨테이너의 `/files`에 마운트된다. 컨테이너 시작 스크립트와 MediaWiki 설정이 여기에 있다.

이 문서는 현재 계약(AS-IS)과 예정된 변경(TO-BE)을 정리한다. 값 자체(비밀값 포함)는 배포 환경이 관리하며 이 저장소에 두지 않는다.

## 환경변수

### goapp (`goapp/app/config/config.go`)

| 변수 | 기본값 | 설명 |
| --- | --- | --- |
| `ENV_FILE` | | 이 경로의 env 파일을 읽어 아래 값을 덮어쓴다(선택) |
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

| 변수 | 사용처 |
| --- | --- |
| `APP_URL` | MediaWiki 설정의 `$wgServer` |
| `REDIS_PERSIST_HOST`, `REDIS_PERSIST_PORT` (없으면 `REDIS_HOST`, `REDIS_PORT`) | ZetaExtension 인증 상태(OTP, 소셜 로그인 연계, `includes/Auth/PersistRedis.php`). goapp이 쓴 토큰을 읽는다 |
| `MW_INSTALL_PATH` | ZetaExtension 유지보수 스크립트. 운영 이미지(`prod`)에서 `/app/w`로 설정 |

### `.env.example`에만 있는 키

`EDITBOT_USERNAME`, `EDITBOT_PASSWORD`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_DEFAULT_REGION`, `AWS_BUCKET`, `AWS_USE_PATH_STYLE_ENDPOINT`는 이 저장소의 코드가 읽지 않는다. 배포 환경이 제공하는 MediaWiki 설정(확장 설정 등)이 읽을 수 있다.

## 배포 환경이 제공하는 파일 (`/files`)

배포 환경은 컨테이너 시작 스크립트를 `/files`에 두고 실행한다. 시작 스크립트가 아래 파일을 제자리로 복사한 뒤 서비스(nginx, php-fpm, goapp 등)를 띄운다. MediaWiki 디렉터리는 운영 이미지(`prod`)에서 `/app/w`, 개발 이미지(`dev`)에서 `/var/www/html`이다.

| 파일 | 복사 위치 | 내용 |
| --- | --- | --- |
| `LocalSettings.php` | MediaWiki 디렉터리 | 환경별 설정. `BaseSettings.php`, 그다음 `/files/ExtensionSettings.php`를 `require` |
| `BaseSettings.php` | MediaWiki 디렉터리 | 공통 설정: 사이트, DB, 캐시, 파일 저장소, 스킨(ZetaSkin). 확장은 로드하지 않는다 |
| `ExtensionSettings.php` | 복사하지 않음 (`/files`에서 바로 `require`) | **모든 확장**(기본 포함, 외부, ZetaExtension)의 `wfLoadExtension`과 설정. 없으면 확장이 하나도 로드되지 않는다 |
| `nginx.conf`, `php-fpm.conf`, `php.ini` | `/etc/nginx`, `/usr/local/etc` | 웹 서버, PHP |
| `supervisord.conf` | `/etc` | 프로세스 구성 (개발 이미지) |
| `dist_ads.txt`, `dist_robots.txt`, `dist_config.js` | `/app/svelte/dist/` | 정적 파일 |
| `SyntaxHighlight.php`, `MsUpload.less`, `mediawiki.skin.defaults.less` | 해당 확장/리소스 경로 | MediaWiki 패치 |
| GA 서비스 계정 JSON | 그대로 | `GA_READER_FILE`이 가리키는 파일 |

### 스킨 상수

ZetaSkin(`mwz/skins/ZetaSkin/includes/SkinZetaSkin.php`)은 다음 PHP 상수를 쓴다. MediaWiki 설정에서 정의되어 있어야 한다.

`ASSET_HASH`, `AVATAR_BASE_URL`, `GA_MEASUREMENT_ID`, `AD_CLIENT`, `AD_SLOTS`

### MediaWiki 확장

- **외부 확장**(git에서 받는, MediaWiki 기본 포함이 아닌 확장): `mwz/extensions.yaml`이 단일 출처다. 목록에 있으면 이미지에 설치된다. 빼려면 주석 처리한다. 확장마다 출처(`repo`/`tag`)를 둔다. 실제로 설치되는 commit은 `mwz/extensions.lock`에 고정되고(`make extensions-lock`), 브랜치(`REL1_43` 등)가 움직여도 lock을 갱신하기 전까지는 같은 commit이 설치된다.
- **ZetaExtension**: 이 저장소의 일부(`mwz/extensions/ZetaExtension`)라 설치 목록에 없다. **ZetaSkin**은 스킨이라 확장과 별도로 MediaWiki 기본 설정이 로드한다.
- **기본 포함 확장**(Cite, VisualEditor 등): 설치할 것이 없다.
- **설치**: 외부 확장은 이미지(`base` 단계)에 들어간다. PHP 의존성은 `hack/mediawiki-composer.lock`.
- **로드와 설정**: 모든 확장(기본 포함, 외부, ZetaExtension)의 `wfLoadExtension`과 설정(`$wg…`)은 배포 환경이 제공하는 `/files/ExtensionSettings.php` 하나가 맡고, `LocalSettings.php`가 `BaseSettings.php` 다음에 이를 `require`한다(위 표). 이 저장소는 설치만 한다. 설치하고 로드하지 않은 확장은 쓰이지 않을 뿐이다.
- 확장 패키지와 로드·설정을 느슨하게 묶어, 운영 고유의 설정을 공개하지 않는다.

## TO-BE

원칙: **MediaWiki 설정 PHP 코드는 이 저장소에서 관리하고 이미지에 넣는다. 환경마다 다른 값과 비밀값은 환경변수로만 받는다.**

- 설정 PHP 구조: `LocalSettings.php`(이 저장소, 진입점과 호출 순서) → `BaseSettings.php`(이 저장소, 모든 환경 공통) → `/files/SiteSettings.php`(배포 환경 주입, 환경별, 확장 제외) → `/files/ExtensionSettings.php`(배포 환경 주입, 모든 확장의 로드와 설정).
- 지금 설정 파일에 들어 있는 값은 `getenv()`로 읽는다. 기존 변수(`DB_*`, `REDIS_PERSIST_*`/`REDIS_VOLATILE_*`([redis.md](redis.md)), `AWS_*`, `AVATAR_BASE_URL`, `GA_MEASUREMENT_ID`, `AD_*`)를 재사용하고, 다음을 추가한다(이름은 확정 전).

  | 변수 | 용도 |
  | --- | --- |
  | `MW_SECRET_KEY`, `MW_UPGRADE_KEY` | `$wgSecretKey`, `$wgUpgradeKey` |
  | `SHELLBOX_SCORE_URL`, `SHELLBOX_SECRET_KEY` | Score 렌더링 |
  | `MW_CDN_SERVERS` | `$wgCdnServers` (쉼표 구분) |

  `.env.example`에 예시 값(`example-db`, `https://example-avatar.example.com` 등)을 둔다.
- 스킨 상수 대신 설정값을 읽어, 배포 환경이 스킨 소스를 고칠 필요가 없게 한다.
