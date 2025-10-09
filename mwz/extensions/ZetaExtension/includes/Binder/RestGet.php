<?php

namespace ZetaExtension\Binder;

use MediaWiki\MediaWikiServices;
use MediaWiki\Rest\SimpleHandler;
use Title;
use Wikimedia\ParamValidator\ParamValidator;

class RestGet extends SimpleHandler
{
    public function getParamSettings()
    {
        return [
            'pageid' => [
                self::PARAM_SOURCE => 'path',
                ParamValidator::PARAM_TYPE => 'string',
                ParamValidator::PARAM_REQUIRED => true,
            ],
        ];
    }

    public function run($pageID)
    {
        if (! is_numeric($pageID) || $pageID < 1) {
            return [];
        }
        $refresh = isset($_GET['refresh']);
        $dbr = MediaWikiServices::getInstance()->getDBLoadBalancer()->getConnection(DB_REPLICA);
        $binderIDs = $dbr->newSelectQueryBuilder()
            ->select(['binder_id'])->from('ldb.binder_pages')
            ->where(['page_id' => $pageID])->fetchFieldValues();
        if (count($binderIDs) == 0) {
            return [];
        }
        if ($refresh) {
            foreach ($binderIDs as $binderID) {
                $wikiPageFactory = MediaWikiServices::getInstance()->getWikiPageFactory();
                $wikiPage = $wikiPageFactory->newFromID($binderID);
                BinderService::syncRelations($wikiPage);
            }
        }
        $res = $dbr->newSelectQueryBuilder()
            ->select(['id', 'data'])->from('ldb.binders')
            ->where(['id' => $binderIDs, 'deleted' => 0, 'enabled' => 1])->fetchResultSet();
        $rows = [];
        foreach ($res as $row) {
            $rows[] = (! $refresh && $row->data) ? unserialize($row->data) : $this->getData($row->id);
        }

        return $rows;
    }

    private function getData($id)
    {
        if (! is_numeric($id)) {
            return -1;
        }
        $t = Title::newFromID($id);
        if (! $t->exists() || $t->getNamespace() != 3000) {
            return -2;
        }
        // /w/api.php?action=parse&format=json&prop=text&pageid=182990
        $json = file_get_contents("http://localhost/w/api.php?action=parse&format=json&prop=text&pageid=$id");
        $parse = json_decode($json, true)['parse'];
        $s = $parse['text']['*'];
        $s = substr($s, strpos($s, '<ul>'));
        $s = substr($s, 0, strrpos($s, '</ul>') + 5);
        $s = str_replace('&amp;', '&', $s);
        $s = preg_replace('/<\/ul>\s*<ul>/', '', $s);
        $s = str_replace('</ul></li><li><ul>', '', $s);
        $s = str_replace('<ul>', '"nodes":[', $s);
        $s = str_replace('</ul>', '],', $s);
        $s = str_replace('<li>', '{', $s);
        $s = str_replace('</li>', '},', $s);
        $s = str_replace('<a href="', '"href":"', $s);
        $s = preg_replace_callback('|>[^<]*</a>|', function ($matches) {
            return str_replace('"', '\"', $matches[0]);
        }, $s);
        $s = str_replace('</a>', '",', $s);
        $s = str_replace(' title="[^"]*"', '', $s);
        $s = str_replace(' class="new"', ',"new":1', $s);
        $s = str_replace(' class="mw-redirect"', ',"redirect":1', $s);
        $s = preg_replace('/\w+="[^"]*"/', '', $s);
        $s = str_replace('>', ',"text":"', $s);
        $s = "[$s]";
        $s = preg_replace('/{([^"^{^}\n]+)/', '{"text":"$1",', $s);
        $s = preg_replace("/,\s*\]/", ']', $s);
        $s = preg_replace("/,\s*}/", '}', $s);
        $s = substr($s, 9, -1);
        $trees = $this->optimizeNodes(json_decode($s, true));

        return [
            'id' => $id,
            'title' => substr($parse['title'], 7),
            'trees' => $trees,
        ];
    }

    private function optimizeNodes($nodes)
    {
        foreach ($nodes as &$node) {
            if (array_key_exists('nodes', $node)) {
                $node['nodes'] = $this->optimizeNodes($node['nodes']);
            }
            if (array_key_exists('new', $node)) {
                continue;
            }
            if (! array_key_exists('title', $node)) {
                continue;
            }
            $t = Title::newFromText($node['title']);
            unset($node['title']);
            if (array_key_exists('redirect', $node)) {
                unset($node['redirect']);
                $t = BinderService::resolveRedirects($t);
            }
            $node['id'] = $t ? $t->getArticleId() : 0;
        }

        return $nodes;
    }
}
