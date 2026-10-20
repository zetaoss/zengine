<?php

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

$wgRightsUrl  = 'https://creativecommons.org/licenses/by-sa/4.0/';
$wgRightsText = '크리에이티브 커먼즈 저작자표시-동일조건변경허락';
$wgRightsIcon = '/w/resources/assets/licenses/cc-by-sa.png';

$wgEmergencyContact = 'admin@mail.zetawiki.com';
$wgPasswordSender   = 'no-reply@mail.zetawiki.com';

// ---- database ----
$wgDBtype     = 'mysql';
$wgDBserver   = getenv('DB_HOST') . ':3306';
$wgDBname     = 'zetawiki';
$wgDBuser     = getenv('DB_USERNAME');
$wgDBpassword = getenv('DB_PASSWORD');

$wgSharedTables[] = 'actor';

// ---- caches ----

$wgObjectCaches['redis-volatile'] = ['class' => 'RedisBagOStuff', 'servers' => [getenv('REDIS_VOLATILE_HOST') . ':6379']];
$wgObjectCaches['redis-persist']  = ['class' => 'RedisBagOStuff', 'servers' => [getenv('REDIS_PERSIST_HOST') . ':6379']];

$wgMainCacheType              = 'redis-volatile';
$wgMessageCacheType           = 'redis-volatile';
$wgLanguageConverterCacheType = 'redis-volatile';
$wgSessionCacheType           = 'redis-persist';

$wgParserCacheType       = CACHE_DB;
$wgParserCacheExpireTime = 86400 * 30;
$wgMemCachedServers      = [];

$wgCacheDirectory                 = "$IP/cache";
$wgLocalisationCacheConf['store'] = 'files';

$wgJobTypeConf['default'] = ['class' => 'JobQueueRedis', 'redisServer' => getenv('REDIS_PERSIST_HOST') . ':6379', 'redisConfig' => [], 'claimTTL' => 3600, 'daemonized' => true];
$wgJobRunRate             = 0;

// ---- CDN ----
$wgUseCdn     = true;
$wgCdnServers = array_filter(explode(',', getenv('MW_CDN_SERVERS')));

// ---- uploads ----
$wgEnableUploads     = true;
$wgUseImageMagick    = true;
$wgUseInstantCommons = true;

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
$wgGroupPermissions['user']['move']                   = false;
$wgGroupPermissions['autoconfirmed']['upload_by_url'] = true;
$wgGroupPermissions['autoconfirmed']['move']          = true;
$wgGroupPermissions['autoconfirmed']['movefile']      = true;
$wgGroupPermissions['autoconfirmed']['move-subpages'] = true;
$wgGroupPermissions['autoconfirmed']['move-rootuserpages'] = true;

$wgDefaultUserOptions['usenewrc']       = 0;
$wgDefaultUserOptions['numberheadings'] = 1;
$wgDefaultUserOptions['usebetatoolbar'] = 1;

// ---- content ----
$wgFixDoubleRedirects = true;
$wgMaxArticleSize     = 512;

$wgEnableMagicLinks = ['ISBN' => true, 'PMID' => true, 'RFC' => true];

$wgFeed                  = false;
$wgShowArchiveThumbnails = false;
$wgAllowExternalImages   = true;
$wgExternalLinkTarget    = '_blank';

// ---- skin ----
$wgDefaultSkin = 'zetaskin';
wfLoadSkin('ZetaSkin');

define('AVATAR_BASE_URL', getenv('AVATAR_BASE_URL'));
define('GA_MEASUREMENT_ID', getenv('GA_MEASUREMENT_ID'));
define('AD_CLIENT', getenv('AD_CLIENT'));
define('AD_SLOTS', array_filter(explode(',', getenv('AD_SLOTS'))));
