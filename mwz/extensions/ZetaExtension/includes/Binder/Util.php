<?php
namespace ZetaExtension\Binder;

use MediaWiki\MediaWikiServices;

class Util
{
    public static function updateBinderPages($wikiPage)
    {
        if ($wikiPage->getTitle()->getNamespace() != 3000) {
            return;
        }
        $binder_id = $wikiPage->getId();
        $text = $wikiPage->getContent()->getText();
        preg_match_all('/\n?(?<stars>\*+)(\s*)\[\[(?<titles>[^\|^\]]+)\|?(?<texts>[^\|^\]]+)?\]\]\s*/', $text, $out);
        $rows = [];
        foreach ($out['titles'] as $i => $title) {
            $t = self::followRedirects(\Title::newFromText($title));
            if (!$t) {
                continue;
            }
            $rows[] = ['binder_id' => $binder_id, 'page_id' => $t->getArticleId()];
        }
        if (count($rows) == 0) {
            return;
        }
        $dbw = MediaWikiServices::getInstance()->getDBLoadBalancer()->getConnection(DB_PRIMARY);
        $dbw->delete('ldb.binder_pages', ['binder_id' => $binder_id]);
        $dbw->insert('ldb.binder_pages', $rows);
        $temp = $dbw->newSelectQueryBuilder()->select('id')->from('ldb.binders')->where(['id' => $binder_id])->fetchField();
        if (!$temp) {
            $dbw->insert('binders', ['id' => $binder_id]);
        }
        $dbw->update('binders', ['data' => null], ['id' => $binder_id]);
    }

    public static function followRedirects($t)
    {
        if (!$t->exists()) {
            return false;
        }
        if ($t->isRedirect()) {
            return self::followRedirects(self::newFromID($t)->followRedirect());
        }
        return $t;
    }

    public static function newFromID($id)
    {
        return MediaWikiServices::getInstance()->getWikiPageFactory()->newFromID($id);
    }
}
