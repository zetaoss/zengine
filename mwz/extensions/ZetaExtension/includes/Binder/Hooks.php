<?php

namespace ZetaExtension\Binder;

use MediaWiki\MediaWikiServices;

class Hooks
{
    public static function onArticleDelete(&$article, &$user, &$reason, &$error)
    {
        if ($error || $article->getTitle()->getNamespace() != 3000) {
            return;
        }
        $binder_id = $article->getId();
        $dbw = MediaWikiServices::getInstance()->getDBLoadBalancer()->getConnection(DB_PRIMARY);
        $dbw->query('UPDATE ldb.binders SET deleted=1 WHERE id=?', $binder_id);
    }

    public static function onArticleUndelete($title, $create, $comment, $oldPageId, $restoredPages)
    {
        if ($title->getNamespace() != 3000) {
            return;
        }
        $binder_id = array_keys($restoredPages)[0] ?? false;
        if (! $binder_id) {
            return;
        }
        $dbw = MediaWikiServices::getInstance()->getDBLoadBalancer()->getConnection(DB_PRIMARY);
        $dbw->query('UPDATE ldb.binders SET deleted=0 WHERE id=?', $binder_id);
        Util::updateBinderPages(Util::newFromId($binder_id));
    }

    public static function onPageSaveComplete($wikiPage, $user, $summary, $flags, $revisionRecord, $editResult)
    {
        if ($wikiPage->getTitle()->getNamespace() != 3000) {
            return;
        }
        Util::updateBinderPages($wikiPage);
    }
}
