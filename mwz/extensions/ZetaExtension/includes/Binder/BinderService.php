<?php

namespace ZetaExtension\Binder;

use MediaWiki\MediaWikiServices;

final class BinderService
{
    private static function dbw()
    {
        return MediaWikiServices::getInstance()->getConnectionProvider()->getPrimaryDatabase();
    }

    private static function dbr()
    {
        return MediaWikiServices::getInstance()->getConnectionProvider()->getReplicaDatabase();
    }

    public static function listBinders(): array
    {
        $res = self::dbr()->newSelectQueryBuilder()
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

        $binderIDs = self::dbr()->newSelectQueryBuilder()
            ->select('binder_id')
            ->from('ldb.binder_pages')
            ->where(['page_id' => $pageId])
            ->fetchFieldValues();

        if (! $binderIDs) {
            return [];
        }

        if ($refresh) {
            $out = [];
            foreach ($binderIDs as $id) {
                $out[] = self::buildAndStoreBinderData($id);
            }

            return $out;
        }

        $res = self::dbr()->newSelectQueryBuilder()
            ->select(['id', 'data'])
            ->from('ldb.binders')
            ->where(['id' => $binderIDs, 'deleted' => 0, 'enabled' => 1])
            ->fetchResultSet();

        $out = [];
        foreach ($res as $row) {
            $data = $row->data ? json_decode($row->data, true) : null;
            if (is_array($data)) {
                $out[] = $data;
            } else {
                $out[] = self::buildAndStoreBinderData((int) $row->id);
            }
        }

        return $out;
    }

    public static function syncRelations(int $binderId): void
    {
        $dbw = self::dbw();
        $res = $dbw->newSelectQueryBuilder()
            ->select(['target_id' => $dbw->buildCoalesce(['p2.page_id', 'p.page_id'])])
            ->from('pagelinks', 'pl')
            ->join('page', 'p', 'p.page_namespace = pl.pl_namespace AND p.page_title = pl.pl_title')
            ->leftJoin('redirect', 'rd', 'rd.rd_from = p.page_id')
            ->leftJoin('page', 'p2', 'p2.page_namespace = rd.rd_namespace AND p2.page_title = rd.rd_title')
            ->where(['pl.pl_from' => $binderId])
            ->fetchResultSet();

        $seen = [];
        $rows = [];
        foreach ($res as $row) {
            $pid = (int) $row->target_id;
            if ($pid > 0 && ! isset($seen[$pid])) {
                $seen[$pid] = true;
                $rows[] = ['binder_id' => $binderId, 'page_id' => $pid];
            }
        }

        $dbw->delete('ldb.binder_pages', ['binder_id' => $binderId]);
        if ($rows) {
            $dbw->insert('ldb.binder_pages', $rows);
        }
        $dbw->upsert('ldb.binders', ['id' => $binderId], ['id'], [], __METHOD__);
    }

    public static function markDeleted(int $binderId): void
    {
        self::dbw()->upsert('ldb.binders', ['id' => $binderId, 'deleted' => 1], ['id'], ['deleted' => 1], __METHOD__);
    }

    public static function unmarkDeletedAndResync(int $binderId): void
    {
        self::dbw()->upsert('ldb.binders', ['id' => $binderId, 'deleted' => 0], ['id'], ['deleted' => 0], __METHOD__);
        self::syncRelations($binderId);
    }

    private static function buildAndStoreBinderData(int $id): array
    {
        $data = self::buildBinderData($id);
        if (! is_array($data)) {
            return [];
        }
        $row = ['id' => $id, 'data' => json_encode($data)];
        self::dbw()->upsert('ldb.binders', $row, ['id'], $row, __METHOD__);

        return $data;
    }

    private static function wikiPageFromId(int $pageId)
    {
        if ($pageId < 1) {
            return null;
        }

        $svc = MediaWikiServices::getInstance();

        if (method_exists($svc, 'getWikiPageFactory')) {
            $page = $svc->getWikiPageFactory()->newFromID($pageId);
        } else {
            $t = \Title::newFromID($pageId);
            $page = $t ? \WikiPage::factory($t) : null;
        }

        if (! $page) {
            return null;
        }
        if (method_exists($page, 'exists') && ! $page->exists()) {
            return null;
        }

        return $page;
    }

    private static function resolveRedirects(\Title $t, int $maxHops = 9): ?\Title
    {
        $svc = MediaWikiServices::getInstance();
        $tf = $svc->getTitleFactory();
        $lookup = $svc->getRedirectLookup();

        while ($maxHops-- > 0) {
            if (! $t->exists() || $t->getId() === 0) {
                return null;
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

    private static function titleToIdByText(string $titleText): int
    {
        $svc = MediaWikiServices::getInstance();
        $tf = $svc->getTitleFactory();

        $t = $tf->newFromText($titleText);
        if (! $t) {
            return 0;
        }

        $t2 = self::resolveRedirects($t);
        if ($t2) {
            $t = $t2;
        }

        return $t->getId();
    }

    private static function buildBinderData(int $id): array
    {
        $page = self::wikiPageFromId($id);
        if (! $page) {
            return [];
        }

        $svc = MediaWikiServices::getInstance();
        $popts = method_exists($svc, 'getParserOptionsFactory')
            ? $svc->getParserOptionsFactory()->newFromAnon()
            : \ParserOptions::newFromAnon();

        $po = method_exists($page, 'getParserOutput') ? $page->getParserOutput($popts) : null;
        if (! $po) {
            return [];
        }

        $html = (string) $po->getText();
        libxml_use_internal_errors(true);
        $dom = new \DOMDocument;
        $dom->loadHTML('<?xml encoding="utf-8" ?>'.$html, LIBXML_NOERROR | LIBXML_NOWARNING);
        libxml_clear_errors();

        $uls = $dom->getElementsByTagName('ul');
        if ($uls->length === 0) {
            return [];
        }

        $title = $page->getTitle()->getText();
        $trees = self::parseUl($uls->item(0));

        return [
            'id' => $id,
            'title' => $title,
            'trees' => $trees,
        ];
    }

    private static function parseUl(\DOMElement $ul): array
    {
        $nodes = [];

        foreach ($ul->childNodes as $li) {
            if ($li->nodeType !== XML_ELEMENT_NODE || $li->nodeName !== 'li') {
                continue;
            }

            $node = ['text' => ''];
            $link = null;

            foreach ($li->childNodes as $child) {
                if ($child->nodeType === XML_ELEMENT_NODE && $child->nodeName === 'a') {
                    $link = $child;
                    break;
                }
            }

            if ($link) {
                $text = trim($link->textContent);
                $href = $link->getAttribute('href');
                $titleAttr = $link->getAttribute('title');

                $node['text'] = $text;
                if ($href !== '') {
                    $node['href'] = $href;
                }

                if ($titleAttr !== '') {
                    $titleText = html_entity_decode($titleAttr, ENT_QUOTES | ENT_HTML5, 'UTF-8');
                    $id = self::titleToIdByText($titleText);
                    if ($id > 0) {
                        $node['id'] = $id;
                    } else {
                        $node['new'] = 1;
                    }
                }
            } else {
                foreach ($li->childNodes as $child) {
                    if ($child->nodeType === XML_TEXT_NODE) {
                        $text = trim($child->textContent);
                        if ($text !== '') {
                            $node['text'] = $text;
                            break;
                        }
                    } elseif ($child->nodeType === XML_ELEMENT_NODE && $child->nodeName === 'ul') {
                        break;
                    }
                }
            }

            foreach ($li->childNodes as $child) {
                if ($child->nodeType === XML_ELEMENT_NODE && $child->nodeName === 'ul') {
                    $node['nodes'] = self::parseUl($child);
                    break;
                }
            }

            $nodes[] = $node;
        }

        return $nodes;
    }
}
