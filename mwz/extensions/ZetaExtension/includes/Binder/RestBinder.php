<?php

namespace ZetaExtension\Binder;

use MediaWiki\Rest\SimpleHandler;
use Wikimedia\ParamValidator\ParamValidator;

class RestBinder extends SimpleHandler
{
    public function getParamSettings()
    {
        return [
            'pageid' => [
                self::PARAM_SOURCE => 'path',
                ParamValidator::PARAM_TYPE => 'integer',
                ParamValidator::PARAM_REQUIRED => true,
            ],
        ];
    }

    public function run($pageid)
    {
        $id = (int) $pageid;
        if ($id < 1) {
            return [];
        }

        $refresh = isset($_GET['refresh']);

        return BinderService::getTreesForPageId($id, $refresh);
    }
}
