<?php

namespace ZetaExtension\Binder;

final class BinderHooks
{
    private const NS_BINDER = 3000;

    public static function onPageSaveComplete($page, $user, $summary, $flags, $revisionRecord, $editResult): void
    {
        if ($page->getNamespace() === self::NS_BINDER) {
            BinderService::syncRelations($page->getId());
        }
    }

    public static function onPageDeleteComplete($page, $reason, $pageID, $revID, $archivedRevisionCount, $user, $timestamp, $logEntry, $archivedFileCount): void
    {
        if ($page->getNamespace() === self::NS_BINDER) {
            BinderService::markDeleted($pageID);
        }
    }

    public static function onPageUndeleteComplete($title, $user, $reason, $oldPageID, $newPageID, $restoredRevisionCount, $logEntry): void
    {
        if ($title->getNamespace() === self::NS_BINDER) {
            BinderService::unmarkDeletedAndResync($newPageID);
        }
    }
}
