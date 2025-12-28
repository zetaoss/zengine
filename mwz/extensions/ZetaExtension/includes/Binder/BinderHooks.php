<?php

namespace ZetaExtension\Binder;

final class BinderHooks
{
    private const NS_BINDER = 3000;

    public static function onPageSaveComplete($wikiPage, $user, $summary, $flags, $revisionRecord, $editResult): void
    {
        if ($wikiPage->getNamespace() === self::NS_BINDER) {
            BinderService::syncRelations($wikiPage->getId());
        }
    }

    public static function onPageDeleteComplete($wikiPage, $user, $reason, $pageId, $deletedRev, $logEntry, $archivedRevisionCount): void
    {
        if ($wikiPage->getNamespace() === self::NS_BINDER) {
            BinderService::markDeleted($pageId);
        }
    }

    public static function onPageUndeleteComplete($title, $user, $reason, $restoredPageId, $restoredRev, $logEntry, $restoredRevisionCount, $created, $restoredPageIds): void
    {
        if ($title->getNamespace() === self::NS_BINDER) {
            BinderService::unmarkDeletedAndResync($restoredPageId);
        }
    }
}
