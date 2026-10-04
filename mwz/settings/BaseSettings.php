<?php

// Settings common to every deployment. Values that differ per deployment and secrets come from
// environment variables (docs/config.md). Extensions are loaded by ExtensionSettings.php.

$zetaEnv = static fn (string $name, string $default = ''): string => trim((string) getenv($name)) ?: $default;
$zetaList = static fn (string $name): array => array_values(array_filter(array_map('trim', explode(',', $zetaEnv($name)))));
// Redis endpoint for a role (PERSIST or VOLATILE); REDIS_HOST/REDIS_PORT serve a role whose variables are unset.
$zetaRedis = static fn (string $role): string => $zetaEnv("REDIS_{$role}_HOST", $zetaEnv('REDIS_HOST')) . ':' . $zetaEnv("REDIS_{$role}_PORT", $zetaEnv('REDIS_PORT', '6379'));

// ---- site ----
$wgSitename = '제타위키';
$wgServer   = $zetaEnv('APP_URL');

$wgArticlePath      = '/wiki/$1';
$wgScriptPath       = '/w';
$wgResourceBasePath = '/w';

$wgLanguageCode  = 'ko';
$wgLocaltimezone = 'Asia/Seoul';

$wgSecretKey  = $zetaEnv('MW_SECRET_KEY');
$wgUpgradeKey = $zetaEnv('MW_UPGRADE_KEY');

$wgAuthenticationTokenVersion = '1';

$wgRightsPage = '';
$wgRightsUrl  = 'https://creativecommons.org/licenses/by-sa/4.0/';
$wgRightsText = '크리에이티브 커먼즈 저작자표시-동일조건변경허락';
$wgRightsIcon = '/w/resources/assets/licenses/cc-by-sa.png';

// ---- email ----
$wgEnableEmail     = true;
$wgEnableUserEmail = true;

$wgEmergencyContact = 'admin@mail.zetawiki.com';
$wgPasswordSender   = 'no-reply@mail.zetawiki.com';

$wgEnotifUserTalk      = false;
$wgEnotifWatchlist     = false;
$wgEmailAuthentication = true;

// ---- database ----
// The database name is fixed: goapp also refers to MediaWiki tables as zetawiki.*.
$wgDBtype     = 'mysql';
$wgDBserver   = $zetaEnv('DB_HOST') . ':' . $zetaEnv('DB_PORT', '3306');
$wgDBname     = 'zetawiki';
$wgDBuser     = $zetaEnv('DB_USERNAME');
$wgDBpassword = $zetaEnv('DB_PASSWORD');

$wgSharedTables[] = 'actor';

// ---- caches and job queue (docs/redis.md) ----
$wgObjectCaches['redis-cache']   = ['class' => 'RedisBagOStuff', 'servers' => [$zetaRedis('VOLATILE')]];
$wgObjectCaches['redis-session'] = ['class' => 'RedisBagOStuff', 'servers' => [$zetaRedis('PERSIST')]];

$wgMainCacheType              = 'redis-cache';
$wgMessageCacheType           = 'redis-cache';
$wgParserCacheType            = 'redis-cache';
$wgLanguageConverterCacheType = 'redis-cache';
$wgSessionCacheType           = 'redis-session';
$wgMemCachedServers           = [];

$wgJobTypeConf['default'] = ['class' => 'JobQueueRedis', 'redisServer' => $zetaRedis('PERSIST'), 'redisConfig' => [], 'claimTTL' => 3600, 'daemonized' => true];
$wgJobRunRate             = 0;

// ---- CDN ----
$wgUseCdn     = true;
$wgCdnServers = $zetaList('MW_CDN_SERVERS');

// ---- uploads ----
$wgEnableUploads     = true;
$wgUseImageMagick    = true;
$wgUseInstantCommons = true;

$wgImageMagickConvertCommand = '/usr/bin/convert';

$wgMaxUploadSize                = 8388608; # 8Mi
$wgAllowCopyUploads             = true;
$wgCopyUploadsFromSpecialUpload = true;

array_push($wgFileExtensions, 'pdf', 'ppt', 'jp2', 'doc', 'docx', 'xls', 'xlsx', 'zip', 'mp3', 'svg', 'diff');

// ---- Shellbox ----
if ($zetaEnv('SHELLBOX_SCORE_URL') !== '') {
    $wgShellboxUrls = ['score' => $zetaEnv('SHELLBOX_SCORE_URL')];
}
$wgShellboxSecretKey = $zetaEnv('SHELLBOX_SECRET_KEY');

// ---- users and permissions ----
$wgAccountCreationThrottle = [['count' => 1, 'seconds' => 86400]];

$wgCookiePrefix = 'wiki';
$wgRememberMe   = 'always';

$wgGroupPermissions['*']['edit']                      = false;
$wgGroupPermissions['autoconfirmed']['upload_by_url'] = true;

$wgDefaultUserOptions['usenewrc']       = 0;
$wgDefaultUserOptions['numberheadings'] = 1;
$wgDefaultUserOptions['usebetatoolbar'] = 1;

// ---- content ----
$wgMaxArticleSize = 512;

$wgDiff3 = '/usr/bin/diff3';

$wgEnableMagicLinks = ['ISBN' => true, 'PMID' => true, 'RFC' => true];

$wgFeed                  = false;
$wgShowArchiveThumbnails = false;
$wgAllowExternalImages   = true;
$wgExternalLinkTarget    = '_blank';

// ---- skin ----
$wgDefaultSkin = 'zetaskin';
wfLoadSkin('ZetaSkin');

// Constants used by ZetaSkin (includes/SkinZetaSkin.php). ASSET_HASH changes when the skin bundle is rebuilt.
define('ASSET_HASH', (string) (@filemtime("$IP/skins/ZetaSkin/dist/app.js") ?: time()));
define('AVATAR_BASE_URL', $zetaEnv('AVATAR_BASE_URL'));
define('GA_MEASUREMENT_ID', $zetaEnv('GA_MEASUREMENT_ID'));
define('AD_CLIENT', $zetaEnv('AD_CLIENT'));
define('AD_SLOTS', json_encode($zetaList('AD_SLOTS')));

unset($zetaEnv, $zetaList, $zetaRedis);
