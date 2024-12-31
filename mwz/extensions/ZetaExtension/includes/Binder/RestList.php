<?php
namespace ZetaExtension\Binder;

use MediaWiki\MediaWikiServices;
use MediaWiki\Rest\SimpleHandler;

class RestList extends SimpleHandler
{
    public function run()
    {
        $dbr = MediaWikiServices::getInstance()->getDBLoadBalancer()->getConnection(DB_PRIMARY);
        $res = $dbr->newSelectQueryBuilder()
            ->select(['A.id', 'B.page_title'])
            ->from('ldb.binders', 'A')
            ->leftJoin('page', 'B', 'A.id=B.page_id')
            ->where(['deleted' => 0, 'enabled' => 1])->fetchResultSet();
        $rows = [];
        foreach ($res as $row) {
            $rows[] = $row;
        }
        return $rows;
    }
}
