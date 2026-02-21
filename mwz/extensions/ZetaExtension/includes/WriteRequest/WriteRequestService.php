<?php

namespace ZetaExtension\WriteRequest;

use MediaWiki\MediaWikiServices;
use Wikimedia\Rdbms\IDatabase;
use Wikimedia\Rdbms\RawSQLValue;

final class WriteRequestService
{
    private static function dbw(): IDatabase
    {
        return MediaWikiServices::getInstance()->getConnectionProvider()->getPrimaryDatabase();
    }

    public static function markDoneIfMatched($wikiPage, $user): void
    {
        $title = $wikiPage->getTitle();
        if (! $title || $title->getNamespace() !== \NS_MAIN || ! $title->canExist()) {
            return;
        }

        $titleText = trim((string) $title->getText());
        if ($titleText === '') {
            return;
        }

        $candidates = array_values(array_unique([
            $titleText,
            str_replace('_', ' ', $titleText),
            str_replace(' ', '_', $titleText),
        ]));

        $writerId = (int) ($user?->getId() ?? 0);
        $now = date('Y-m-d H:i:s');

        self::dbw()->newUpdateQueryBuilder()
            ->update('ldb.write_requests')
            ->set([
                'writer_id' => $writerId > 0 ? $writerId : 0,
                'writer_name' => $writerId > 0 ? (string) $user->getName() : 'Unknown',
                'writed_at' => $now,
                'updated_at' => $now,
            ])
            ->where([
                'title' => $candidates,
                'writer_id' => -1,
            ])
            ->caller(__METHOD__)
            ->execute();
    }

    public static function incrementNotMatchHit(string $titleText): void
    {
        $titleText = trim($titleText);
        if ($titleText === '') {
            return;
        }
        $now = date('Y-m-d H:i:s');

        self::dbw()->upsert(
            'ldb.not_matches',
            [
                'title' => $titleText,
                'hit' => 1,
                'created_at' => $now,
                'updated_at' => $now,
            ],
            ['title'],
            [
                'hit' => new RawSQLValue('hit+1'),
                'updated_at' => $now,
            ],
            __METHOD__
        );
    }
}
