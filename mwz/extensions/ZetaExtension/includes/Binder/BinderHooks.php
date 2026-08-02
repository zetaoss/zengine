<?php

namespace ZetaExtension\Binder;

use ZetaExtension\WriteRequest\WriteRequestService;

final class BinderHooks
{
    public static function onPageSaveComplete($wikiPage, $user, $summary, $flags, $revisionRecord, $editResult): void
    {
        if ($wikiPage->getNamespace() === NS_BINDER) {
            $pageId = (int) $wikiPage->getId();
            $row = BinderService::ensureBinder($pageId);
        }

        WriteRequestService::markDoneIfMatched($wikiPage, $user);
    }

    public static function onPageDeleteComplete($wikiPage, $user, $reason, $pageId, $deletedRev, $logEntry, $archivedRevisionCount): void
    {
        if ($wikiPage->getNamespace() === NS_BINDER) {
            BinderService::deleteBinder($pageId);
        }
    }

    public static function onPageUndeleteComplete($title, $user, $reason, $restoredPageId, $restoredRev, $logEntry, $restoredRevisionCount, $created, $restoredPageIds): void
    {
        if ($title->getNamespace() === NS_BINDER) {
            $row = BinderService::ensureBinder($restoredPageId);
        }
    }
}
