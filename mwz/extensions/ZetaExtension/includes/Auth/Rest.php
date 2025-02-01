<?php

namespace ZetaExtension\Auth;

use MediaWiki\MediaWikiServices;
use MediaWiki\Rest\SimpleHandler;
use Predis\Client;
use Title;
use User;
use Wikimedia\ParamValidator\ParamValidator;

class Rest extends SimpleHandler
{
    public function run()
    {
        $params = $this->getValidatedParams();
        $code = $params['code'];
        $username = $params['username'];

        // validate code
        $redis = new Client(['host' => getenv('REDIS_HOST')]);
        if (! $redis) {
            return ['status' => 'error', 'error' => 'internal error'];
        }
        $social_id = $redis->get("zetawiki_cache_:code:$code");
        if (! $social_id || ! is_numeric($social_id) || $social_id < 1) {
            return ['status' => 'error', 'error' => 'invalid code'];
        }

        // check creatable
        $newusername = Title::makeTitleSafe(NS_USER, ucfirst($username));
        $newuser = User::newFromName($newusername->getText(), 'creatable');
        if (! $newuser->isAnon()) {
            return ['status' => 'error', 'error' => 'not createable username'];
        }

        // add new user
        $newuser->addToDatabase();
        $user_id = $newuser->getId();

        // update user_social
        $dbw = MediaWikiServices::getInstance()->getDBLoadBalancer()->getConnection(DB_PRIMARY);
        $dbw->update('user_social', ['user_id' => $user_id], ['id' => $social_id], __METHOD__);

        return ['status' => 'success', 'data' => 'ok'];
    }

    public function getParamSettings()
    {
        return [
            'code' => [
                self::PARAM_SOURCE => 'path',
                ParamValidator::PARAM_TYPE => 'string',
                ParamValidator::PARAM_REQUIRED => true,
            ],
            'username' => [
                self::PARAM_SOURCE => 'query',
                ParamValidator::PARAM_TYPE => 'string',
                ParamValidator::PARAM_REQUIRED => true,
            ],
        ];
    }
}
