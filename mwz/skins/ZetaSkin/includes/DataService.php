<?php

namespace ZetaSkin;

use MediaWiki\MediaWikiServices;
use ObjectCache;

class DataService
{
    public static function getUserAvatar(int $userID): ?array
    {
        if ($userID == 0) {
            return null;
        }

        $cache = ObjectCache::getLocalClusterInstance();
        $key = "userAvatar:$userID";
        $cached = $cache->get($key);
        if ($cached !== false) {
            return $cached;
        }

        $dbr = MediaWikiServices::getInstance()
            ->getDBLoadBalancer()
            ->getConnection(DB_REPLICA);

        $row = $dbr->newSelectQueryBuilder()
            ->select(['user_name', 't', 'ghash'])
            ->from('user', 'A')
            ->leftJoin('ldb.profiles', 'B', 'A.user_id = B.user_id')
            ->where(['A.user_id' => $userID])
            ->caller(__METHOD__)
            ->fetchRow();

        if (! $row) {
            return null;
        }

        $userAvatar = [
            'id' => $userID,
            'name' => $row->user_name,
            't' => $row->t,
            'ghash' => $row->ghash,
        ];

        $cache->set($key, $userAvatar);

        return $userAvatar;
    }

    public static function getContributors($title): array
    {
        $url = 'http://localhost/w/api.php?format=json&action=query&prop=contributors&titles='.rawurlencode($title);

        $contents = @file_get_contents($url);
        if ($contents === false) {
            return [];
        }

        $data = json_decode($contents, true);
        if (! is_array($data)) {
            return [];
        }

        $pages = $data['query']['pages'] ?? [];
        if (empty($pages)) {
            return [];
        }

        $contributors = array_pop($pages)['contributors'] ?? [];
        if (! is_array($contributors)) {
            return [];
        }

        return array_map(
            fn ($user) => self::getUserAvatar((int) ($user['userid'] ?? 0)),
            $contributors
        );
    }

    public static function getBinders(int $pageid)
    {
        $url = "http://localhost/w/rest.php/binder/$pageid";
        $contents = @file_get_contents($url);
        if ($contents === false) {
            return [];
        }

        return json_decode($contents, true);
    }
}
