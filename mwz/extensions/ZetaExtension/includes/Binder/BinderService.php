<?php

namespace ZetaExtension\Binder;

use MediaWiki\MediaWikiServices;

final class BinderService
{
    private static function dbw()
    {
        $svc = MediaWikiServices::getInstance();

        return method_exists($svc, 'getConnectionProvider')
            ? $svc->getConnectionProvider()->getPrimaryDatabase()
            : $svc->getDBLoadBalancer()->getConnection(DB_PRIMARY);
    }

    private static function dbr()
    {
        $svc = MediaWikiServices::getInstance();

        return method_exists($svc, 'getConnectionProvider')
            ? $svc->getConnectionProvider()->getReplicaDatabase()
            : $svc->getDBLoadBalancer()->getConnection(DB_REPLICA);
    }

    private static function wikiPageFromId(int $pageId)
    {
        $svc = MediaWikiServices::getInstance();
        if (method_exists($svc, 'getWikiPageFactory')) {
            return $svc->getWikiPageFactory()->newFromID($pageId);
        }
        $t = \Title::newFromID($pageId);

        return $t ? \WikiPage::factory($t) : null;
    }

    public static function listBinders(): array
    {
        $dbr = self::dbr();
        $res = $dbr->newSelectQueryBuilder()
            ->select(['A.id', 'B.page_title'])
            ->from('ldb.binders', 'A')
            ->leftJoin('page', 'B', 'A.id = B.page_id')
            ->where(['A.deleted' => 0, 'A.enabled' => 1])
            ->fetchResultSet();

        $rows = [];
        foreach ($res as $row) {
            $rows[] = ['id' => (int) $row->id, 'title' => (string) $row->page_title];
        }

        return $rows;
    }

    public static function getBindersForPageId(int $pageId, bool $refresh = false): array
    {
        if ($pageId < 1) {
            return [];
        }
        $dbr = self::dbr();
        $binderIDs = $dbr->newSelectQueryBuilder()
            ->select('binder_id')->from('ldb.binder_pages')
            ->where(['page_id' => $pageId])->fetchFieldValues();
        if (! $binderIDs) {
            return [];
        }

        if ($refresh) {
            foreach ($binderIDs as $bid) {
                $wp = self::wikiPageFromId((int) $bid);
                if ($wp) {
                    self::syncRelations($wp);
                }
            }
        }

        $res = $dbr->newSelectQueryBuilder()
            ->select(['id', 'data'])->from('ldb.binders')
            ->where(['id' => $binderIDs, 'deleted' => 0, 'enabled' => 1])
            ->fetchResultSet();

        $rows = [];
        foreach ($res as $row) {
            $rows[] = (! $refresh && $row->data) ? unserialize($row->data) : self::buildBinderData((int) $row->id);
        }

        return $rows;
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
        $wp = self::wikiPageFromId($binderId);
        if ($wp) {
            self::syncRelations($wp);
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
        $tf = $svc->getTitleFactory();

        $lookup = $svc->getRedirectLookup();
        $seen = [];
        while ($maxHops-- > 0) {
            if (! $t->exists()) {
                return null;
            }
            $pid = $t->getId();
            if ($pid > 0) {
                if (isset($seen[$pid])) {
                    return $t;
                } $seen[$pid] = true;
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

    private static function buildBinderData(int $id): array|int
    {
        $json = @file_get_contents("http://localhost/w/api.php?action=parse&format=json&prop=text&pageid={$id}");
        if ($json === false) {
            return -2;
        }

        $parse = json_decode($json, true)['parse'] ?? null;
        if (! $parse || ! isset($parse['text']['*'])) {
            return -2;
        }

        $s = $parse['text']['*'];
        $s = substr($s, strpos($s, '<ul>'));
        $s = substr($s, 0, strrpos($s, '</ul>') + 5);
        $s = str_replace('&amp;', '&', $s);
        $s = preg_replace('/<\/ul>\s*<ul>/', '', $s);
        $s = str_replace('</ul></li><li><ul>', '', $s);
        $s = str_replace('<ul>', '"nodes":[', $s);
        $s = str_replace('</ul>', '],', $s);
        $s = str_replace('<li>', '{', $s);
        $s = str_replace('</li>', '},', $s);
        $s = str_replace('<a href="', '"href":"', $s);
        $s = preg_replace_callback('|>[^<]*</a>|', fn ($m) => str_replace('"', '\"', $m[0]), $s);
        $s = str_replace('</a>', '",', $s);
        $s = str_replace(' title="[^"]*"', '', $s);
        $s = str_replace(' class="new"', ',"new":1', $s);
        $s = str_replace(' class="mw-redirect"', ',"redirect":1', $s);
        $s = preg_replace('/\w+="[^"]*"/', '', $s);
        $s = str_replace('>', ',"text":"', $s);
        $s = "[$s]";
        $s = preg_replace('/{([^"^{^}\n]+)/', '{"text":"$1",', $s);
        $s = preg_replace("/,\s*\]/", ']', $s);
        $s = preg_replace("/,\s*}/", '}', $s);
        $s = substr($s, 9, -1);

        $trees = self::optimizeNodes(json_decode($s, true));

        return [
            'id' => $id,
            'title' => substr((string) $parse['title'], 7),
            'trees' => $trees,
        ];
    }

    private static function optimizeNodes($nodes)
    {
        foreach ($nodes as &$node) {
            if (array_key_exists('nodes', $node)) {
                $node['nodes'] = self::optimizeNodes($node['nodes']);
            }
            if (array_key_exists('new', $node)) {
                continue;
            }
            if (! array_key_exists('title', $node)) {
                continue;
            }
            $t = \Title::newFromText($node['title']);
            unset($node['title']);
            if (array_key_exists('redirect', $node)) {
                unset($node['redirect']);
                $t = self::resolveRedirects($t);
            }
            $node['id'] = $t ? $t->getArticleId() : 0;
        }

        return $nodes;
    }
}
