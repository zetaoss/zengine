<?php

namespace ZetaExtension\Binder;

use MediaWiki\MediaWikiServices;
use Wikimedia\Rdbms\IDatabase;

final class BinderService
{
    private const MAX_REDIRECT_HOPS = 9;

    private static function dbw(): IDatabase
    {
        return MediaWikiServices::getInstance()->getConnectionProvider()->getPrimaryDatabase();
    }

    private static function dbr(): IDatabase
    {
        return MediaWikiServices::getInstance()->getConnectionProvider()->getReplicaDatabase();
    }

    public static function listBinders(): array
    {
        $res = self::dbr()->newSelectQueryBuilder()
            ->select(['A.id', 'B.page_title', 'A.docs', 'A.links', 'A.title_doc'])
            ->from('ldb.binders', 'A')
            ->leftJoin('page', 'B', 'A.id = B.page_id')
            ->caller(__METHOD__)
            ->fetchResultSet();

        $rows = [];
        foreach ($res as $row) {
            $rows[] = [
                'id' => (int) $row->id,
                'title' => (string) ($row->page_title ?? ''),
                'docs' => (int) ($row->docs ?? 0),
                'links' => (int) ($row->links ?? 0),
                'title_doc' => (string) ($row->title_doc ?? ''),
            ];
        }

        return $rows;
    }

    public static function getTreesForPageId(int $pageId): array
    {
        if ($pageId < 1) {
            return [];
        }

        $caches = self::dbr()->newSelectQueryBuilder()
            ->select('B.cache')
            ->distinct()
            ->from('ldb.binder_pages', 'BP')
            ->join('ldb.binders', 'B', 'B.id = BP.binder_id')
            ->where([
                'BP.page_id' => $pageId,
                'B.enabled' => 1,
            ])
            ->caller(__METHOD__)
            ->fetchFieldValues();

        $out = [];
        foreach ($caches as $cache) {
            $tree = self::decodeCache($cache);
            if ($tree !== []) {
                $out[] = $tree;
            }
        }

        return $out;
    }

    public static function rebuildBinder(int $binderId): ?array
    {
        return self::ensureBinder($binderId);
    }

    public static function ensureBinder(int $binderId): ?array
    {
        if ($binderId < 1) {
            return null;
        }

        $realBinderId = self::resolveRealBinderId($binderId);
        if ($realBinderId < 1) {
            return null;
        }

        if ($realBinderId !== $binderId) {
            self::deleteBinder($binderId);
        }

        $tree = self::buildTree($realBinderId);
        if ($tree === null) {
            return null;
        }

        $dbw = self::dbw();
        $dbw->startAtomic(__METHOD__);
        try {
            self::syncRelations($realBinderId, self::collectNodeReferences($tree['nodes'] ?? []));
            $row = self::storeBinder($realBinderId, $tree);
            $dbw->endAtomic(__METHOD__);
        } catch (\Throwable $e) {
            $dbw->cancelAtomic(__METHOD__);
            throw $e;
        }

        return $row;
    }

    private static function syncRelations(int $binderId, array $references): void
    {
        $dbw = self::dbw();

        $seen = [];
        $rows = [];
        foreach ($references as $reference) {
            $key = $reference['page_namespace'].':'.$reference['page_title'];
            if (isset($seen[$key])) {
                continue;
            }

            $seen[$key] = true;
            $rows[] = [
                'binder_id' => $binderId,
                'page_namespace' => $reference['page_namespace'],
                'page_title' => $reference['page_title'],
                'page_id' => $reference['page_id'],
            ];
        }

        $dbw->delete('ldb.binder_pages', ['binder_id' => $binderId]);
        if ($rows) {
            $dbw->insert('ldb.binder_pages', $rows);
        }
    }

    private static function storeBinder(int $binderId, array $tree): array
    {
        $dbw = self::dbw();
        $now = $dbw->timestamp();
        $row = [
            'id' => $binderId,
            'cache' => json_encode($tree, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES),
            'docs' => self::countDocs($tree['nodes'] ?? []),
            'links' => self::countLinks($tree['nodes'] ?? []),
            'title_doc' => self::findTitleDoc($tree['nodes'] ?? []),
            'created_at' => $now,
            'updated_at' => $now,
        ];
        $updateRow = $row;
        unset($updateRow['id']);
        unset($updateRow['created_at']);
        $dbw->upsert('ldb.binders', $row, ['id'], $updateRow, __METHOD__);

        return $row;
    }

    public static function deleteBinder(int $binderId): void
    {
        $dbw = self::dbw();
        $dbw->delete('ldb.binder_pages', ['binder_id' => $binderId]);
        $dbw->delete('ldb.binders', ['id' => $binderId]);
    }

    public static function refreshByTitle($title, int $pageId = 0, bool $force = false): void
    {
        $dbw = self::dbw();
        $relations = $dbw->newSelectQueryBuilder()
            ->select(['binder_id', 'page_id'])
            ->from('ldb.binder_pages')
            ->where([
                'page_namespace' => $title->getNamespace(),
                'page_title' => $title->getDBkey(),
            ])
            ->caller(__METHOD__)
            ->fetchResultSet();

        $binderIds = [];
        foreach ($relations as $relation) {
            $indexedPageId = $relation->page_id !== null ? (int) $relation->page_id : 0;
            if ($force || $pageId < 1 || $indexedPageId !== $pageId) {
                $binderIds[] = (int) $relation->binder_id;
            }
        }

        foreach (array_unique($binderIds) as $binderId) {
            self::ensureBinder($binderId);
        }
    }

    private static function decodeCache($raw): array
    {
        if (! is_string($raw) || $raw === '') {
            return [];
        }

        $tree = json_decode($raw, true);

        return is_array($tree) ? $tree : [];
    }

    private static function wikiPageFromId(int $pageId)
    {
        if ($pageId < 1) {
            return null;
        }

        $svc = MediaWikiServices::getInstance();
        $page = $svc->getWikiPageFactory()->newFromID($pageId);

        if (! $page || ! $page->exists()) {
            return null;
        }

        return $page;
    }

    private static function resolveRedirects(\Title $t, int $maxHops = self::MAX_REDIRECT_HOPS): ?\Title
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

    private static function resolveRealBinderId(int $binderId): int
    {
        $page = self::wikiPageFromId($binderId);
        if (! $page) {
            return 0;
        }

        $sourceTitle = $page->getTitle();
        $targetTitle = self::resolveRedirects($sourceTitle) ?: $sourceTitle;

        return (int) $targetTitle->getId();
    }

    private static function buildTree(int $id): ?array
    {
        $page = self::wikiPageFromId($id);
        if (! $page) {
            return null;
        }

        $sourceTitle = $page->getTitle();
        $targetTitle = self::resolveRedirects($sourceTitle) ?: $sourceTitle;
        $parsePage = $page;

        if ($targetTitle->getId() > 0 && $targetTitle->getId() !== $sourceTitle->getId()) {
            $resolvedPage = self::wikiPageFromId((int) $targetTitle->getId());
            if ($resolvedPage) {
                $parsePage = $resolvedPage;
            }
        }

        $title = str_replace('_', ' ', $targetTitle->getText());
        $html = $parsePage->getParserOutput()->getText();
        libxml_use_internal_errors(true);
        $dom = new \DOMDocument;
        $dom->loadHTML('<?xml encoding="utf-8" ?>'.$html, LIBXML_NOERROR | LIBXML_NOWARNING);
        libxml_clear_errors();

        $uls = $dom->getElementsByTagName('ul');
        if ($uls->length === 0) {
            return [
                'id' => $id,
                'text' => $title,
                'nodes' => [],
            ];
        }

        $nodes = self::parseUl($uls->item(0));

        return [
            'id' => $id,
            'text' => $title,
            'nodes' => $nodes,
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
                $titleAttr = self::linkTitleText($link);

                $node['text'] = $text;
                if ($href !== '') {
                    $node['href'] = $href;
                }

                if ($titleAttr !== '') {
                    $titleText = html_entity_decode($titleAttr, ENT_QUOTES | ENT_HTML5, 'UTF-8');
                    $node['title'] = $titleText;
                    $targetId = self::titleToIdByText($titleText);
                    $linkTitle = MediaWikiServices::getInstance()->getTitleFactory()->newFromText($titleText);
                    if ($targetId > 0) {
                        $node['id'] = $targetId;
                        if ($linkTitle) {
                            $node['href'] = $linkTitle->getLocalURL();
                        }
                    } else {
                        $node['new'] = 1;
                        if ($linkTitle) {
                            $node['href'] = $linkTitle->getLocalURL([
                                'action' => 'edit',
                                'redlink' => 1,
                            ]);
                        }
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

    private static function linkTitleText(\DOMElement $link): string
    {
        if (in_array('new', preg_split('/\s+/', $link->getAttribute('class')) ?: [], true)) {
            $query = parse_url(html_entity_decode($link->getAttribute('href'), ENT_QUOTES | ENT_HTML5, 'UTF-8'), PHP_URL_QUERY);
            if (is_string($query)) {
                parse_str($query, $params);
                if (isset($params['title']) && is_string($params['title'])) {
                    return str_replace('_', ' ', $params['title']);
                }
            }
        }

        return $link->getAttribute('title');
    }

    private static function countDocs(array $nodes): int
    {
        $count = 0;

        foreach ($nodes as $node) {
            if (! is_array($node)) {
                continue;
            }

            if (isset($node['id']) && (int) $node['id'] > 0) {
                $count++;
            }

            if (isset($node['nodes']) && is_array($node['nodes'])) {
                $count += self::countDocs($node['nodes']);
            }
        }

        return $count;
    }

    private static function collectNodeReferences(array $nodes): array
    {
        $references = [];

        foreach ($nodes as $node) {
            if (! is_array($node)) {
                continue;
            }

            if (isset($node['title']) && is_string($node['title'])) {
                foreach (self::titleToReferences($node['title']) as $reference) {
                    $references[] = $reference;
                }
            }

            if (isset($node['nodes']) && is_array($node['nodes'])) {
                foreach (self::collectNodeReferences($node['nodes']) as $childReference) {
                    $references[] = $childReference;
                }
            }
        }

        return $references;
    }

    private static function titleToReferences(string $titleText): array
    {
        $services = MediaWikiServices::getInstance();
        $titleFactory = $services->getTitleFactory();
        $redirectLookup = $services->getRedirectLookup();
        $title = $titleFactory->newFromText($titleText);
        if (! $title) {
            return [];
        }

        $dependencies = [];
        $seen = [];
        $resolvedPageId = null;
        $redirectHops = 0;
        while (true) {
            $key = $title->getNamespace().':'.$title->getDBkey();
            if (isset($seen[$key])) {
                break;
            }

            $seen[$key] = true;
            $dependencies[] = $title;
            if (! $title->exists() || $title->getId() === 0) {
                break;
            }

            $target = $redirectLookup->getRedirectTarget($title);
            if (! $target) {
                $resolvedPageId = (int) $title->getId();
                break;
            }
            if ($redirectHops >= self::MAX_REDIRECT_HOPS) {
                break;
            }

            $next = $titleFactory->newFromLinkTarget($target);
            if (! $next) {
                break;
            }
            $title = $next;
            $redirectHops++;
        }

        $references = [];
        foreach ($dependencies as $dependency) {
            $references[] = [
                'page_namespace' => $dependency->getNamespace(),
                'page_title' => $dependency->getDBkey(),
                'page_id' => $resolvedPageId ?: null,
            ];
        }

        return $references;
    }

    private static function countLinks(array $nodes): int
    {
        $count = 0;

        foreach ($nodes as $node) {
            if (! is_array($node)) {
                continue;
            }

            if (isset($node['href']) && is_string($node['href']) && $node['href'] !== '') {
                $count++;
            }

            if (isset($node['nodes']) && is_array($node['nodes'])) {
                $count += self::countLinks($node['nodes']);
            }
        }

        return $count;
    }

    private static function findTitleDoc(array $nodes): string
    {
        foreach ($nodes as $node) {
            if (! is_array($node)) {
                continue;
            }

            if (isset($node['id']) && (int) $node['id'] > 0) {
                if (isset($node['title']) && is_string($node['title']) && $node['title'] !== '') {
                    return $node['title'];
                }

                return isset($node['text']) && is_string($node['text']) ? $node['text'] : '';
            }

            if (isset($node['nodes']) && is_array($node['nodes'])) {
                $titleDoc = self::findTitleDoc($node['nodes']);
                if ($titleDoc !== '') {
                    return $titleDoc;
                }
            }
        }

        return '';
    }
}
