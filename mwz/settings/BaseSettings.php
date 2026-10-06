<?php

// Settings common to every deployment. Values that differ per deployment and secrets come from
// environment variables (docs/config.md). Extensions are loaded by ExtensionSettings.php.

// ---- site ----
$wgSitename = '제타위키';
$wgServer   = getenv('APP_URL');

$wgArticlePath      = '/wiki/$1';
$wgScriptPath       = '/w';
$wgResourceBasePath = '/w';

$wgLanguageCode  = 'ko';
$wgLocaltimezone = 'Asia/Seoul';

$wgSecretKey  = getenv('MW_SECRET_KEY');
$wgUpgradeKey = getenv('MW_UPGRADE_KEY');

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
$wgDBserver   = getenv('DB_HOST') . ':' . getenv('DB_PORT');
$wgDBname     = 'zetawiki';
$wgDBuser     = getenv('DB_USERNAME');
$wgDBpassword = getenv('DB_PASSWORD');

$wgSharedTables[] = 'actor';

// ---- caches and job queue (docs/redis.md) ----
$redisPort = 6379;
$wgObjectCaches['redis-volatile'] = ['class' => 'RedisBagOStuff', 'servers' => [getenv('REDIS_VOLATILE_HOST') . ':' . $redisPort]];
$wgObjectCaches['redis-persist']  = ['class' => 'RedisBagOStuff', 'servers' => [getenv('REDIS_PERSIST_HOST') . ':' . $redisPort]];

$wgMainCacheType              = 'redis-volatile';
$wgMessageCacheType           = 'redis-volatile';
$wgParserCacheType            = CACHE_DB;
$wgParserCacheExpireTime      = 86400 * 30;
$wgLanguageConverterCacheType = 'redis-volatile';
$wgSessionCacheType           = 'redis-persist';
$wgMemCachedServers           = [];

$wgJobTypeConf['default'] = ['class' => 'JobQueueRedis', 'redisServer' => getenv('REDIS_PERSIST_HOST') . ':' . $redisPort, 'redisConfig' => [], 'claimTTL' => 3600, 'daemonized' => true];
$wgJobRunRate             = 0;

// ---- CDN ----
$wgUseCdn     = true;
$wgCdnServers = array_filter(explode(',', getenv('MW_CDN_SERVERS')));

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
$wgShellboxUrls      = ['score' => getenv('SHELLBOX_SCORE_URL')];
$wgShellboxSecretKey = getenv('SHELLBOX_SECRET_KEY');

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
define('AVATAR_BASE_URL', getenv('AVATAR_BASE_URL'));
define('GA_MEASUREMENT_ID', getenv('GA_MEASUREMENT_ID'));
define('AD_CLIENT', getenv('AD_CLIENT'));
define('AD_SLOTS', json_encode(array_filter(explode(',', getenv('AD_SLOTS')))));
