<?php

namespace ZetaExtension\Binder;

use MediaWiki\MediaWikiServices;

final class BinderService
{
    private static function dbw()
    {
        $svc = MediaWikiServices::getInstance();
        if (method_exists($svc, 'getConnectionProvider')) {
            return $svc->getConnectionProvider()->getPrimaryDatabase();
        }

        return $svc->getDBLoadBalancer()->getConnection(DB_PRIMARY);
    }

    public static function syncRelations($page): void
    {
        $binderId = $page->getId();
        if ($binderId <= 0) {
            return;
        }

        $text = (string) $page->getContent()?->getText();
        if ($text === '') {
            self::writeRelations($binderId, []);

            return;
        }

        if (! preg_match_all('/\[\[\s*(?<title>[^\|\]\n#]+)(?:#[^\|\]]*)?(?:\|[^\]]*)?\s*\]\]/u', $text, $matches, PREG_SET_ORDER)) {
            self::writeRelations($binderId, []);

            return;
        }

        $svc = MediaWikiServices::getInstance();
        $tf = $svc->getTitleFactory();
        $rows = [];
        $seen = [];

        foreach ($matches as $m) {
            $raw = trim($m['title']);
            if ($raw === '') {
                continue;
            }
            $t = $tf->newFromText($raw);
            if (! $t) {
                continue;
            }
            $t = self::resolveRedirects($t);
            if (! $t) {
                continue;
            }
            $pid = $t->getId();
            if ($pid > 0 && ! isset($seen[$pid])) {
                $seen[$pid] = true;
                $rows[] = ['binder_id' => $binderId, 'page_id' => $pid];
            }
        }

        self::writeRelations($binderId, $rows);
    }

    public static function markDeleted(int $binderId): void
    {
        self::dbw()->upsert('ldb.binders', ['id' => $binderId, 'deleted' => 1], ['id'], ['deleted' => 1], __METHOD__);
    }

    public static function unmarkDeletedAndResync(int $binderId): void
    {
        self::dbw()->upsert('ldb.binders', ['id' => $binderId, 'deleted' => 0], ['id'], ['deleted' => 0], __METHOD__);
        $page = MediaWikiServices::getInstance()->getWikiPageFactory()->newFromID($binderId);
        if ($page) {
            self::syncRelations($page);
        }
    }

    private static function writeRelations(int $binderId, array $rows): void
    {
        $dbw = self::dbw();
        $dbw->delete('ldb.binder_pages', ['binder_id' => $binderId]);
        if ($rows) {
            $dbw->insert('ldb.binder_pages', $rows);
        }
        $dbw->upsert('ldb.binders', ['id' => $binderId], ['id'], ['id' => $binderId], __METHOD__);
    }

    private static function resolveRedirects(\Title $t, int $maxHops = 9): ?\Title
    {
        $svc = MediaWikiServices::getInstance();
        $lookup = $svc->getRedirectLookup();
        $tf = $svc->getTitleFactory();
        $seen = [];

        while ($maxHops-- > 0) {
            if (! $t->exists()) {
                return null;
            }
            $pid = $t->getId();
            if ($pid > 0) {
                if (isset($seen[$pid])) {
                    return $t;
                }
                $seen[$pid] = true;
            }
            $target = $lookup->getRedirectTarget($t);
            if (! $target) {
                return $t;
            }
            $t = $tf->newFromLinkTarget($target);
            if (! $t) {
                return null;
            }
        }

        return $t;
    }
}
