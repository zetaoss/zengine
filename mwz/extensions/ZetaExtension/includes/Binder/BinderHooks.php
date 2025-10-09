<?php

namespace ZetaExtension\Binder;

final class BinderHooks
{
    private const NS_BINDER = 3000;

    public static function onPageSaveComplete($page, $user, $summary, $flags, $revisionRecord, $editResult): void
    {
        if ($page->getNamespace() !== self::NS_BINDER) {
            return;
        }
        BinderService::syncRelations($page);
    }

    public static function onPageDeleteComplete($page, $reason, $pageID, $revID, $archivedRevisionCount, $user, $timestamp, $logEntry, $archivedFileCount): void
    {
        if ($page->getNamespace() !== self::NS_BINDER) {
            return;
        }
        $binderId = $pageID ?? $page->getId();
        if ($binderId > 0) {
            BinderService::markDeleted((int) $binderId);
        }
    }

    public static function onPageUndeleteComplete($title, $user, $reason, $oldPageID, $newPageID, $restoredRevisionCount, $logEntry): void
    {
        if ($title->getNamespace() !== self::NS_BINDER) {
            return;
        }
        $binderId = $newPageID ?: $title->getId();
        if ($binderId > 0) {
            BinderService::unmarkDeletedAndResync((int) $binderId);
        }
    }
}
