<?php

namespace ZetaExtension\WriteRequest;

final class SearchHooks
{
    public static function onSpecialSearchNogomatch(&$title): void
    {
        if (! $title || ! $title->canExist()) {
            return;
        }
        if ($title->exists()) {
            return;
        }

        WriteRequestService::incrementNotMatchHit((string) $title->getText());
    }
}
