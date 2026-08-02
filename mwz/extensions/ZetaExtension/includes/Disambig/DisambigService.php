<?php

namespace ZetaExtension\Disambig;

use JobSpecification;
use MediaWiki\MediaWikiServices;
use Wikimedia\Rdbms\IDatabase;

final class DisambigService
{
    private static function dbw(): IDatabase
    {
        return MediaWikiServices::getInstance()->getConnectionProvider()->getPrimaryDatabase();
    }

    public static function ensureDisambig(int $disambigId): ?array
    {
        if ($disambigId < 1) {
            return null;
        }

        $list = self::buildList($disambigId);
        if ($list === null) {
            return null;
        }

        $dbw = self::dbw();
        $dbw->startAtomic(__METHOD__);
        try {
            self::syncRelations($disambigId, $list['references']);
            unset($list['references']);
            $row = self::storeDisambig($disambigId, $list);
            $dbw->endAtomic(__METHOD__);
        } catch (\Throwable $e) {
            $dbw->cancelAtomic(__METHOD__);
            throw $e;
        }

        return $row;
    }

    public static function deleteDisambig(int $disambigId): void
    {
        $dbw = self::dbw();
        $dbw->startAtomic(__METHOD__);
        try {
            $dbw->delete('ldb.disambig_pages', ['disambig_id' => $disambigId], __METHOD__);
            $dbw->delete('ldb.disambigs', ['id' => $disambigId], __METHOD__);
            $dbw->endAtomic(__METHOD__);
        } catch (\Throwable $e) {
            $dbw->cancelAtomic(__METHOD__);
            throw $e;
        }
    }

    public static function enqueueLegacyDisambigCleanupJobs(
        int $disambigId,
        int $userId,
        string $userName,
        string $baseTitle,
        string $disambigTitle
    ): void {
        $dbw = self::dbw();
        $relations = $dbw->newSelectQueryBuilder()
            ->select('page_title')
            ->from('ldb.disambig_pages')
            ->where(['disambig_id' => $disambigId])
            ->caller(__METHOD__)
            ->fetchResultSet();

        $services = MediaWikiServices::getInstance();
        $titleFactory = $services->getTitleFactory();
        $jobParams = [
            'user_id' => $userId,
            'user_name' => $userName,
            'disambig_title' => $disambigTitle,
        ];
        $jobOptions = [
            'removeDuplicates' => true,
            'removeDuplicatesIgnoreParams' => ['requestId'],
        ];
        $jobs = [];
        foreach ($relations as $relation) {
            $title = $titleFactory->makeTitle(NS_MAIN, (string) $relation->page_title);
            $jobs[] = new JobSpecification(
                CleanupLegacyDisambigJob::TYPE,
                $jobParams,
                $jobOptions,
                $title
            );
        }

        $cache = $dbw->newSelectQueryBuilder()
            ->select('cache')
            ->from('ldb.disambigs')
            ->where(['id' => $disambigId])
            ->caller(__METHOD__)
            ->fetchField();
        $list = is_string($cache) ? json_decode($cache, true) : null;
        $nodes = is_array($list) && is_array($list['nodes'] ?? null) ? $list['nodes'] : [];
        $firstNode = is_array($nodes[0] ?? null) ? $nodes[0] : [];
        $firstTitle = (string) ($firstNode['title'] ?? '');
        $basePage = $titleFactory->makeTitle(NS_MAIN, $baseTitle);
        $baseWikiPage = $services->getWikiPageFactory()->newFromTitle($basePage);
        $firstPage = $firstTitle !== '' ? $titleFactory->newFromText($firstTitle) : null;
        $hasBaseNode = false;
        foreach ($nodes as $node) {
            if (! is_array($node)) {
                continue;
            }
            $nodePage = $titleFactory->newFromText((string) ($node['title'] ?? ''));
            if ($nodePage !== null && $nodePage->getPrefixedDBkey() === $basePage->getPrefixedDBkey()) {
                $hasBaseNode = true;
                break;
            }
        }
        if (
            $firstPage !== null
            && ! $hasBaseNode
            && (! $baseWikiPage->exists() || $baseWikiPage->isRedirect())
            && $basePage->getPrefixedDBkey() !== $firstPage->getPrefixedDBkey()
        ) {
            $jobs[] = new JobSpecification(
                CleanupLegacyDisambigJob::TYPE,
                $jobParams + ['redirect_target' => $firstPage->getPrefixedText()],
                $jobOptions,
                $basePage
            );
        }

        if ($jobs) {
            MediaWikiServices::getInstance()->getJobQueueGroup()->lazyPush($jobs);
        }
    }

    public static function attachPage($title, int $pageId): void
    {
        if ($pageId < 1 || $title->getNamespace() !== NS_MAIN) {
            return;
        }

        $dbw = self::dbw();
        $disambigIds = $dbw->newSelectQueryBuilder()
            ->select('disambig_id')
            ->from('ldb.disambig_pages')
            ->where([
                'page_title' => $title->getDBkey(),
                'page_id' => null,
            ])
            ->caller(__METHOD__)
            ->fetchFieldValues();
        if (! $disambigIds) {
            return;
        }

        foreach (array_unique($disambigIds) as $disambigId) {
            self::ensureDisambig((int) $disambigId);
        }
    }

    public static function clearPageId(int $pageId): void
    {
        if ($pageId < 1) {
            return;
        }

        $dbw = self::dbw();
        $disambigIds = $dbw->newSelectQueryBuilder()
            ->select('disambig_id')
            ->from('ldb.disambig_pages')
            ->where(['page_id' => $pageId])
            ->caller(__METHOD__)
            ->fetchFieldValues();
        if (! $disambigIds) {
            return;
        }

        foreach (array_unique($disambigIds) as $disambigId) {
            self::ensureDisambig((int) $disambigId);
        }
    }

    private static function syncRelations(int $disambigId, array $references): void
    {
        $rows = [];
        $seen = [];
        foreach ($references as $reference) {
            $key = $reference['page_title'];
            if (isset($seen[$key])) {
                continue;
            }

            $seen[$key] = true;
            $rows[] = [
                'disambig_id' => $disambigId,
                'page_title' => $reference['page_title'],
                'page_id' => $reference['page_id'],
            ];
        }

        $dbw = self::dbw();
        $dbw->delete('ldb.disambig_pages', ['disambig_id' => $disambigId], __METHOD__);
        if ($rows) {
            $dbw->insert('ldb.disambig_pages', $rows, __METHOD__);
        }
    }

    private static function storeDisambig(int $disambigId, array $list): array
    {
        $dbw = self::dbw();
        $now = $dbw->timestamp();
        $row = [
            'id' => $disambigId,
            'cache' => json_encode($list, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES),
            'entries' => count($list['nodes'] ?? []),
            'created_at' => $now,
            'updated_at' => $now,
        ];
        $updateRow = $row;
        unset($updateRow['id'], $updateRow['created_at']);
        $dbw->upsert('ldb.disambigs', $row, ['id'], $updateRow, __METHOD__);

        return $row;
    }

    private static function buildList(int $disambigId): ?array
    {
        $services = MediaWikiServices::getInstance();
        $page = $services->getWikiPageFactory()->newFromID($disambigId);
        if (! $page || ! $page->exists() || $page->getNamespace() !== NS_DISAMBIG) {
            return null;
        }

        $parserOutput = $page->getParserOutput();
        if (! $parserOutput) {
            return null;
        }

        $html = $parserOutput->getText();
        libxml_use_internal_errors(true);
        $dom = new \DOMDocument;
        $dom->loadHTML('<?xml encoding="utf-8" ?>'.$html, LIBXML_NOERROR | LIBXML_NOWARNING);
        libxml_clear_errors();

        $nodes = [];
        $references = [];
        $uls = $dom->getElementsByTagName('ul');
        if ($uls->length > 0) {
            foreach ($uls->item(0)->childNodes as $li) {
                if (! $li instanceof \DOMElement || $li->nodeName !== 'li') {
                    continue;
                }

                $link = self::findDirectLink($li);
                if (! $link) {
                    continue;
                }

                $titleText = html_entity_decode(self::linkTitleText($link), ENT_QUOTES | ENT_HTML5, 'UTF-8');
                $title = $services->getTitleFactory()->newFromText($titleText);
                if (! $title || $title->getNamespace() !== NS_MAIN) {
                    continue;
                }

                $pageId = (int) $title->getId();
                $node = [
                    'text' => trim($link->textContent),
                    'title' => $title->getPrefixedText(),
                    'href' => $pageId > 0
                        ? $title->getLocalURL()
                        : $title->getLocalURL(['action' => 'edit', 'redlink' => 1]),
                    'description' => self::extractDescription($li, $link),
                ];
                if ($pageId > 0) {
                    $node['id'] = $pageId;
                } else {
                    $node['new'] = 1;
                }
                $nodes[] = $node;
                $references[] = [
                    'page_title' => $title->getDBkey(),
                    'page_id' => $pageId > 0 ? $pageId : null,
                ];
            }
        }

        return [
            'id' => $disambigId,
            'text' => $page->getTitle()->getText(),
            'nodes' => $nodes,
            'references' => $references,
        ];
    }

    private static function findDirectLink(\DOMElement $li): ?\DOMElement
    {
        foreach ($li->childNodes as $child) {
            if ($child instanceof \DOMElement && $child->nodeName === 'a') {
                return $child;
            }
        }

        return null;
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

    private static function extractDescription(\DOMElement $li, \DOMElement $link): string
    {
        $parts = [];
        $afterLink = false;
        foreach ($li->childNodes as $child) {
            if ($child === $link) {
                $afterLink = true;
                continue;
            }
            if (! $afterLink || ($child instanceof \DOMElement && $child->nodeName === 'ul')) {
                continue;
            }
            $parts[] = $child->textContent;
        }

        $description = trim(implode('', $parts));

        return trim((string) preg_replace('/^\s*[-–—]\s*/u', '', $description));
    }
}
