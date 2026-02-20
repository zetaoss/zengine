<?php

namespace ZetaSkin;

use MediaWiki\MediaWikiServices;

class PageDataProvider
{
    private static $dbr;

    public static function fetchBinders(int $pageId): array
    {
        if ($pageId < 1) {
            return [];
        }

        $res = self::dbr()->newSelectQueryBuilder()
            ->distinct()
            ->select(['B.data'])
            ->from('ldb.binder_pages', 'BP')
            ->join('ldb.binders', 'B', 'B.id = BP.binder_id')
            ->where([
                'BP.page_id' => $pageId,
                'B.deleted' => 0,
                'B.enabled' => 1,
            ])
            ->fetchResultSet();

        $binders = [];
        foreach ($res as $row) {
            $data = $row->data ? json_decode($row->data, true) : null;
            if (is_array($data) && $data) {
                $binders[] = $data;
            }
        }

        return $binders;
    }

    public static function fetchContributors(int $pageId): array
    {
        if ($pageId < 1) {
            return [];
        }

        $rows = self::dbr()->newSelectQueryBuilder()
            ->select([
                'AC.actor_user',
                'AC.actor_name',
                'last_rev' => 'MAX(R.rev_timestamp)',
            ])
            ->from('revision', 'R')
            ->join('actor', 'AC', 'R.rev_actor = AC.actor_id')
            ->where([
                'R.rev_page' => $pageId,
                'AC.actor_user > 0',
            ])
            ->groupBy(['AC.actor_user', 'AC.actor_name'])
            ->orderBy('last_rev', 'DESC')
            ->limit(10)
            ->fetchResultSet();

        $contributors = [];
        foreach ($rows as $row) {
            $id = (int) ($row->actor_user ?? 0);
            if ($id < 1) {
                continue;
            }
            $contributors[] = ['id' => $id, 'name' => $row->actor_name ?? ''];
        }

        return $contributors;
    }

    private static function dbr()
    {
        if (! self::$dbr) {
            self::$dbr = MediaWikiServices::getInstance()->getConnectionProvider()->getReplicaDatabase();
        }

        return self::$dbr;
    }
}
