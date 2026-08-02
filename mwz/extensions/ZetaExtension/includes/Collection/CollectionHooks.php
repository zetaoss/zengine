<?php

namespace ZetaExtension\Collection;

use ZetaExtension\Binder\BinderService;
use ZetaExtension\Disambig\DisambigService;
use ZetaExtension\WriteRequest\WriteRequestService;

final class CollectionHooks
{
    public static function onPageSaveComplete($wikiPage, $user, $summary, $flags, $revisionRecord, $editResult): void
    {
        $pageId = (int) $wikiPage->getId();
        $title = $wikiPage->getTitle();
        BinderService::attachPage($title, $pageId);
        DisambigService::attachPage($title, $pageId);

        if ($wikiPage->getNamespace() === NS_BINDER) {
            BinderService::ensureBinder($pageId);
        } elseif ($wikiPage->getNamespace() === NS_DISAMBIG) {
            $disambig = DisambigService::ensureDisambig($pageId);
            if ($disambig !== null) {
                DisambigService::enqueueLegacyDisambigCleanupJobs(
                    $pageId,
                    (int) $user->getId(),
                    $user->getName(),
                    $title->getText(),
                    $title->getPrefixedText()
                );
            }
        }

        WriteRequestService::markDoneIfMatched($wikiPage, $user);
    }

    public static function onPageDeleteComplete($page, $deleter, $reason, $pageId, $deletedRev, $logEntry, $archivedRevisionCount): void
    {
        if ($page->getNamespace() === NS_BINDER) {
            BinderService::deleteBinder($pageId);
        } elseif ($page->getNamespace() === NS_DISAMBIG) {
            DisambigService::deleteDisambig($pageId);
        }

        BinderService::clearPageId($pageId);
        DisambigService::clearPageId($pageId);
    }

    public static function onPageUndeleteComplete($page, $restorer, $reason, $restoredRev, $logEntry, $restoredRevisionCount, $created, $restoredPageIds): void
    {
        $pageId = (int) $page->getId();
        BinderService::attachPage($page, $pageId);
        DisambigService::attachPage($page, $pageId);

        if ($page->getNamespace() === NS_BINDER) {
            BinderService::ensureBinder($pageId);
        } elseif ($page->getNamespace() === NS_DISAMBIG) {
            DisambigService::ensureDisambig($pageId);
        }
    }

    public static function onPageMoveComplete($old, $new, $user, $pageId, $redirectId, $reason, $revision): void
    {
        BinderService::clearPageId($pageId);
        DisambigService::clearPageId($pageId);
        BinderService::attachPage($new, $pageId);
        DisambigService::attachPage($new, $pageId);

        if ($redirectId > 0) {
            BinderService::attachPage($old, $redirectId);
            DisambigService::attachPage($old, $redirectId);
        }

        if ($old->getNamespace() === NS_BINDER && $new->getNamespace() !== NS_BINDER) {
            BinderService::deleteBinder($pageId);
        } elseif ($new->getNamespace() === NS_BINDER) {
            BinderService::ensureBinder($pageId);
        }

        if ($old->getNamespace() === NS_DISAMBIG && $new->getNamespace() !== NS_DISAMBIG) {
            DisambigService::deleteDisambig($pageId);
        } elseif ($new->getNamespace() === NS_DISAMBIG) {
            DisambigService::ensureDisambig($pageId);
        }
    }
}
