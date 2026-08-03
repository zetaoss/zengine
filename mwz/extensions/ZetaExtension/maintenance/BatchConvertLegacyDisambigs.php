<?php

use MediaWiki\CommentStore\CommentStoreComment;
use MediaWiki\Content\ContentHandler;
use MediaWiki\Content\TextContent;
use MediaWiki\MediaWikiServices;
use MediaWiki\Revision\SlotRecord;
use ZetaExtension\Disambig\CleanupLegacyDisambigJob;
use ZetaExtension\Disambig\DisambigService;

require getenv('MW_INSTALL_PATH') . '/maintenance/Maintenance.php';

class BatchConvertLegacyDisambigs extends Maintenance
{
    public function __construct()
    {
        parent::__construct();

        $this->addDescription('Batch convert legacy {{다른뜻}} templates into Disambig pages or strip them if conditions are not met.');
        $this->addOption('dry-run', 'Perform a dry run without modifying any pages or database entries.');
        $this->addOption('user', 'User name to perform edits under.', false, true);
        $this->addOption('limit', 'Maximum number of base titles to process.', false, true);
    }

    public function execute(): void
    {
        $isDryRun = (bool)$this->getOption('dry-run', false);
        $userName = (string)$this->getOption('user', 'Jmnote bot');
        $limit = (int)$this->getOption('limit', 0);

        $services = MediaWikiServices::getInstance();
        $userFactory = $services->getUserFactory();
        $user = $userFactory->newFromName($userName);
        if (!$user || $user->getName() === '') {
            $this->fatalError("User '{$userName}' not found.");
        }

        if ($isDryRun) {
            $this->output("=== DRY RUN MODE ===\n");
        }

        $dbr = $services->getConnectionProvider()->getReplicaDatabase();
        $wikiPageFactory = $services->getWikiPageFactory();
        $titleFactory = $services->getTitleFactory();

        // 1. Find linktarget IDs for legacy templates
        $ltIds = $dbr->newSelectQueryBuilder()
            ->select('lt_id')
            ->from('linktarget')
            ->where(['lt_namespace' => NS_TEMPLATE, 'lt_title' => ['다른뜻', '다른_뜻']])
            ->caller(__METHOD__)
            ->fetchFieldValues();

        $pageIds = [];

        if ($ltIds !== []) {
            // Fast indexed lookup from templatelinks
            $tlRows = $dbr->newSelectQueryBuilder()
                ->select('tl_from')
                ->from('templatelinks')
                ->where(['tl_target_id' => array_map('intval', $ltIds)])
                ->caller(__METHOD__)
                ->fetchResultSet();
            foreach ($tlRows as $row) {
                $pageIds[(int)$row->tl_from] = true;
            }
        }

        if ($pageIds === []) {
            $this->output("No pages found with legacy {{다른뜻}} template in templatelinks.\n");
            return;
        }

        $pageIdList = array_keys($pageIds);

        // Fetch valid NS_MAIN non-redirect pages
        $pageRows = $dbr->newSelectQueryBuilder()
            ->select(['page_id', 'page_title'])
            ->from('page')
            ->where(['page_id' => $pageIdList, 'page_namespace' => NS_MAIN, 'page_is_redirect' => 0])
            ->caller(__METHOD__)
            ->fetchResultSet();

        $pagesWithLegacyTemplate = [];
        $baseTitleMap = []; // baseTitle => list of page titles

        foreach ($pageRows as $row) {
            $titleText = str_replace('_', ' ', (string)$row->page_title);
            $baseTitle = $this->extractBaseTitle($titleText);

            $pagesWithLegacyTemplate[] = [
                'id' => (int)$row->page_id,
                'title' => $titleText,
                'baseTitle' => $baseTitle,
            ];

            if (!isset($baseTitleMap[$baseTitle])) {
                $baseTitleMap[$baseTitle] = [];
            }
            $baseTitleMap[$baseTitle][] = $titleText;
        }

        $totalBaseTitles = count($baseTitleMap);
        $this->output(sprintf("Found %d page(s) with legacy template across %d base title(s).\n", count($pagesWithLegacyTemplate), $totalBaseTitles));

        if ($totalBaseTitles === 0) {
            return;
        }

        $processedBaseTitles = 0;
        $createdDisambigs = 0;
        $strippedOnlyCount = 0;
        $errors = [];

        foreach ($baseTitleMap as $baseTitle => $sourcePageTitles) {
            if ($limit > 0 && $processedBaseTitles >= $limit) {
                $this->output(sprintf("Reached limit of %d base titles.\n", $limit));
                break;
            }

            $processedBaseTitles++;
            $this->output(sprintf("[%d/%d] Base title: '%s'\n", $processedBaseTitles, $totalBaseTitles, $baseTitle));

            // Find all candidate pages sharing this base title
            $candidates = $this->findCandidatePages($baseTitle);

            // Extract labels/extra candidates from legacy templates
            $labels = [];
            foreach ($sourcePageTitles as $srcTitleStr) {
                $srcTitleObj = $titleFactory->newFromText($srcTitleStr);
                if (!$srcTitleObj || !$srcTitleObj->exists()) {
                    continue;
                }
                $srcWikiPage = $wikiPageFactory->newFromTitle($srcTitleObj);
                $content = $srcWikiPage->getContent();
                if ($content instanceof TextContent) {
                    $parsedMeanings = $this->parseOtherMeanings($content->getText(), $baseTitle);
                    foreach ($parsedMeanings as $m) {
                        $mTitle = $m['title'];
                        if (!in_array($mTitle, $candidates, true) && $this->isValidCandidateTitle($mTitle, $baseTitle)) {
                            $mTitleObj = $titleFactory->newFromText($mTitle);
                            if ($mTitleObj && $mTitleObj->getNamespace() === NS_MAIN) {
                                // Include uncreated pages or existing non-redirect pages
                                if (!$mTitleObj->exists()) {
                                    $candidates[] = $mTitle;
                                } else {
                                    $mWikiPage = $wikiPageFactory->newFromTitle($mTitleObj);
                                    if (!$mWikiPage->isRedirect()) {
                                        $candidates[] = $mTitle;
                                    }
                                }
                            }
                        }
                        if ($m['label'] !== '' && !isset($labels[$mTitle])) {
                            $labels[$mTitle] = $m['label'];
                        }
                    }
                }
            }

            // Deduplicate candidates
            $candidates = array_values(array_unique($candidates));

            // Identify uncreated candidates
            $uncreatedCandidates = [];
            foreach ($candidates as $candTitleStr) {
                $candTitleObj = $titleFactory->newFromText($candTitleStr);
                if ($candTitleObj && ($candTitleObj->getId() === 0 || !$candTitleObj->exists())) {
                    $uncreatedCandidates[] = $candTitleStr;
                }
            }

            // Format candidate strings for log output (uncreated candidates in red)
            $formattedCandidates = [];
            foreach ($candidates as $candTitleStr) {
                $lbl = $labels[$candTitleStr] ?? '';
                $dispStr = ($lbl !== '' && $lbl !== $candTitleStr) ? "{$candTitleStr}|{$lbl}" : $candTitleStr;
                if (in_array($candTitleStr, $uncreatedCandidates, true)) {
                    $formattedCandidates[] = "\033[31m" . $dispStr . "\033[0m";
                } else {
                    $formattedCandidates[] = $dispStr;
                }
            }

            // Check if Disambig:<baseTitle> already exists
            $disambigTitleObj = $titleFactory->makeTitle(NS_DISAMBIG, $baseTitle);
            $disambigExists = $disambigTitleObj->exists();

            if (count($candidates) >= 2 || $disambigExists) {
                // Condition MET: Generate or update Disambig document and strip templates
                $this->output(sprintf("  -> Candidate count: %d (>=2). Generating Disambig and stripping templates.\n", count($candidates)));
                if ($uncreatedCandidates !== []) {
                    $this->output(sprintf("  -> \033[31mUncreated candidate(s) (%d): %s\033[0m\n", count($uncreatedCandidates), implode(', ', $uncreatedCandidates)));
                }

                if (!$isDryRun) {
                    try {
                        // Create Disambig page if it doesn't exist
                        if (!$disambigExists) {
                            $wikitextLines = [];
                            foreach ($candidates as $candTitle) {
                                $lbl = $labels[$candTitle] ?? '';
                                if ($lbl !== '' && $lbl !== $candTitle) {
                                    $wikitextLines[] = sprintf("* [[%s|%s]]", $candTitle, $lbl);
                                } else {
                                    $wikitextLines[] = sprintf("* [[%s]]", $candTitle);
                                }
                            }
                            $newWikitext = implode("\n", $wikitextLines);

                            $disambigWikiPage = $wikiPageFactory->newFromTitle($disambigTitleObj);
                            $updater = $disambigWikiPage->newPageUpdater($user);
                            $updater->setContent(
                                SlotRecord::MAIN,
                                ContentHandler::makeContent($newWikitext, $disambigTitleObj, CONTENT_MODEL_WIKITEXT)
                            );
                            $updater->saveRevision(
                                CommentStoreComment::newUnsavedComment("일괄 마이그레이션: 동음이의어 문서 자동 생성"),
                                EDIT_MINOR
                            );

                            if (!$updater->wasSuccessful()) {
                                throw new RuntimeException("Failed to save Disambig page '{$disambigTitleObj->getPrefixedText()}'");
                            }

                            $createdDisambigs++;
                            $this->output(sprintf("  [CREATED] %s\n", $disambigTitleObj->getPrefixedText()));
                        }

                        // Synchronize Disambig relations & cache
                        $disambigId = (int)$disambigTitleObj->getId();
                        if ($disambigId > 0) {
                            DisambigService::ensureDisambig($disambigId);
                        }

                        // Strip legacy templates from candidate pages
                        foreach ($candidates as $candTitleStr) {
                            $this->stripTemplateFromPage($candTitleStr, $user);
                        }
                    } catch (Throwable $e) {
                        $errors[] = sprintf("BaseTitle '%s': %s", $baseTitle, $e->getMessage());
                        $this->output(sprintf("  [ERROR] %s\n", $e->getMessage()));
                    }
                } else {
                    $this->output(sprintf("  [DRY-RUN] Would create Disambig '%s' with candidates: %s\n", $disambigTitleObj->getPrefixedText(), implode(', ', $formattedCandidates)));
                    $this->output(sprintf("  [DRY-RUN] Would strip templates from: %s\n", implode(', ', $sourcePageTitles)));
                }
            } else {
                // Condition NOT MET (< 2 candidates): Only strip legacy templates
                $this->output(sprintf("  -> Candidate count: %d (<2). Stripping template only.\n", count($candidates)));
                if ($uncreatedCandidates !== []) {
                    $this->output(sprintf("  -> \033[31mUncreated candidate(s) (%d): %s\033[0m\n", count($uncreatedCandidates), implode(', ', $uncreatedCandidates)));
                }

                if (!$isDryRun) {
                    try {
                        foreach ($sourcePageTitles as $srcTitleStr) {
                            $this->stripTemplateFromPage($srcTitleStr, $user);
                        }
                        $strippedOnlyCount++;
                    } catch (Throwable $e) {
                        $errors[] = sprintf("BaseTitle '%s': %s", $baseTitle, $e->getMessage());
                        $this->output(sprintf("  [ERROR] %s\n", $e->getMessage()));
                    }
                } else {
                    $this->output(sprintf("  [DRY-RUN] Would strip templates from: %s\n", implode(', ', $sourcePageTitles)));
                }
            }
        }

        $this->output("\n=== SUMMARY ===\n");
        $this->output(sprintf("Processed Base Titles: %d\n", $processedBaseTitles));
        $this->output(sprintf("Disambig Pages Created: %d\n", $createdDisambigs));
        $this->output(sprintf("Template Only Stripped: %d\n", $strippedOnlyCount));

        if ($errors !== []) {
            $this->output("Errors encountered:\n- " . implode("\n- ", $errors) . "\n");
            $this->fatalError(sprintf("%d base title(s) failed during batch migration.", count($errors)));
        }
    }

    private function extractBaseTitle(string $titleText): string
    {
        return preg_match('/^(.+?)(?: )?\([^()]+\)$/u', $titleText, $matches)
            ? trim($matches[1])
            : trim($titleText);
    }

    private function isValidCandidateTitle(string $titleText, string $baseTitle): bool
    {
        $title = trim($titleText);
        if ($title === $baseTitle) {
            return true;
        }

        if (mb_strpos($title, $baseTitle) !== 0) {
            return false;
        }

        $suffix = mb_substr($title, mb_strlen($baseTitle));
        if (!preg_match('/^ ?\([^()]+\)$/u', $suffix)) {
            return false;
        }

        $qualifier = trim(mb_substr($suffix, mb_strpos($suffix, '(') + 1, -1));
        return $qualifier !== '';
    }

    private function findCandidatePages(string $baseTitle): array
    {
        $services = MediaWikiServices::getInstance();
        $dbr = $services->getConnectionProvider()->getReplicaDatabase();
        $baseDBKey = str_replace(' ', '_', $baseTitle);

        $likeSpace = $dbr->buildLike($baseDBKey . '_(', $dbr->anyString());
        $likeNoSpace = $dbr->buildLike($baseDBKey . '(', $dbr->anyString());

        $rows = $dbr->newSelectQueryBuilder()
            ->select('page_title')
            ->from('page')
            ->where(['page_namespace' => NS_MAIN, 'page_is_redirect' => 0])
            ->andWhere(
                $dbr->makeList([
                    'page_title' => $baseDBKey,
                    'page_title ' . $likeSpace,
                    'page_title ' . $likeNoSpace,
                ], LIST_OR)
            )
            ->caller(__METHOD__)
            ->fetchResultSet();

        $candidates = [];
        foreach ($rows as $row) {
            $tStr = str_replace('_', ' ', (string)$row->page_title);
            if ($this->isValidCandidateTitle($tStr, $baseTitle)) {
                $candidates[] = $tStr;
            }
        }

        return $candidates;
    }

    private function parseOtherMeanings(string $wikitext, string $baseTitle): array
    {
        $meanings = [];
        $withoutComments = preg_replace('/<!--[\s\S]*?-->/u', '', $wikitext);
        $pattern = '/\{\{\s*다른[\s_]*뜻\s*\|([^{}]*?)\}\}/iu';

        if (preg_match_all($pattern, $withoutComments, $matches)) {
            foreach ($matches[1] as $paramStr) {
                $params = array_map('trim', explode('|', $paramStr));
                $first = $params[0] ?? '';
                $second = $params[1] ?? '';

                $targetTitle = $second !== '' ? $second : $first;
                $label = $second !== '' ? $first : '';

                if ($targetTitle !== '' && $this->isValidCandidateTitle($targetTitle, $baseTitle)) {
                    $meanings[] = [
                        'title' => $targetTitle,
                        'label' => $label,
                    ];
                }
            }
        }

        return $meanings;
    }

    private function stripTemplateFromPage(string $titleStr, \MediaWiki\User\User $user): void
    {
        $services = MediaWikiServices::getInstance();
        $titleObj = $services->getTitleFactory()->newFromText($titleStr);
        if (!$titleObj || !$titleObj->exists()) {
            return;
        }

        $wikiPage = $services->getWikiPageFactory()->newFromTitle($titleObj);
        if ($wikiPage->isRedirect()) {
            return;
        }

        $updater = $wikiPage->newPageUpdater($user);
        $parentRev = $updater->grabParentRevision();
        if (!$parentRev) {
            return;
        }

        $content = $parentRev->getContent(SlotRecord::MAIN);
        if (!$content instanceof TextContent) {
            return;
        }

        $currentWikitext = $content->getText();
        $cleanedWikitext = CleanupLegacyDisambigJob::stripLegacyDisambigTemplates($currentWikitext);

        if ($cleanedWikitext === $currentWikitext) {
            return;
        }

        $updater->setContent(
            SlotRecord::MAIN,
            ContentHandler::makeContent($cleanedWikitext, $titleObj, $content->getModel())
        );

        $updater->saveRevision(
            CommentStoreComment::newUnsavedComment("일괄 마이그레이션: {{다른뜻}} 레거시 틀 정리"),
            EDIT_MINOR
        );

        if (!$updater->wasSuccessful()) {
            throw new RuntimeException("Failed to strip legacy template from page '{$titleObj->getPrefixedText()}'");
        }
    }
}

$maintClass = BatchConvertLegacyDisambigs::class;
require RUN_MAINTENANCE_IF_MAIN;
