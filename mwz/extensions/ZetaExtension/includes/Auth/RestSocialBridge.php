<?php

namespace ZetaExtension\Auth;

use MediaWiki\MediaWikiServices;
use MediaWiki\Rest\Response;
use MediaWiki\Rest\SimpleHandler;
use Redis;
use Wikimedia\ParamValidator\ParamValidator;

class RestSocialBridge extends SimpleHandler
{
    public function execute()
    {
        $qp = $this->getRequest()->getQueryParams();
        $token = (string) ($qp['token'] ?? '');

        if ($token === '') {
            return $this->getResponseFactory()->createJson([
                'status' => 'error',
                'message' => 'missing token',
            ], 400);
        }

        $payload = $this->popPayload($token);
        if (! $payload) {
            return $this->getResponseFactory()->createJson([
                'status' => 'error',
                'message' => 'invalid token',
            ], 403);
        }

        $userId = (int) ($payload['user_id'] ?? 0);
        if ($userId < 1) {
            return $this->getResponseFactory()->createJson([
                'status' => 'error',
                'message' => 'invalid user_id',
            ], 403);
        }

        $services = MediaWikiServices::getInstance();
        $user = $services->getUserFactory()->newFromId($userId);
        if (! $user || ! $user->getId()) {
            return $this->getResponseFactory()->createJson([
                'status' => 'error',
                'message' => 'user not found',
            ], 404);
        }

        $session = $this->getSession();
        $session->setUser($user);
        $session->setRememberUser(true);
        $session->persist();

        $returnto = (string) ($payload['returnto'] ?? '');
        $location = $this->buildReturntoLocation($returnto);

        $res = new Response;
        $res->setStatus(303);
        $res->setHeader('Location', $location);

        return $res;
    }

    public function getAllowedParams()
    {
        return [
            'token' => [
                ParamValidator::PARAM_TYPE => 'string',
                ParamValidator::PARAM_REQUIRED => true,
            ],
        ];
    }

    private function buildReturntoLocation(string $returnto): string
    {
        $returnto = trim($returnto);
        if ($returnto === '') {
            return '/';
        }

        $returnto = str_replace(' ', '_', $returnto);
        $encoded = implode('/', array_map(
            static fn ($seg) => rawurlencode($seg),
            explode('/', $returnto)
        ));

        return '/wiki/'.$encoded;
    }

    private function popPayload(string $token): ?array
    {
        $redis = new Redis;
        $redis->connect(getenv('REDIS_HOST'));

        $raw = $redis->getDel("mwbridge:{$token}");
        if (! is_string($raw) || $raw === '') {
            return null;
        }

        $data = json_decode($raw, true);

        return is_array($data) ? $data : null;
    }
}
