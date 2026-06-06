<?php

namespace ZetaExtension\AIEdit;

use MediaWiki\MediaWikiServices;
use MediaWiki\Rest\SimpleHandler;

class RestPreference extends SimpleHandler
{
    private function json(array $data, int $status = 200)
    {
        $res = $this->getResponseFactory()->createJson($data);
        $res->setStatus($status);

        return $res;
    }

    private function readJsonBody(): ?array
    {
        $body = $this->getRequest()->getBody();
        $raw = $body ? $body->getContents() : '';
        if (! is_string($raw) || $raw === '') {
            return null;
        }
        $data = json_decode($raw, true);

        return is_array($data) ? $data : null;
    }

    public function execute()
    {
        $data = $this->readJsonBody();
        if (! $data) {
            return $this->json(['status' => 'error', 'message' => 'invalid json body'], 400);
        }

        $secret = (string) ($data['secret'] ?? '');
        $internalSecret = getenv('INTERNAL_SECRET_KEY');
        if ($secret === '' || $secret !== $internalSecret) {
            return $this->json(['status' => 'error', 'message' => 'unauthorized'], 403);
        }

        $userId = (int) ($data['user_id'] ?? 0);
        if ($userId <= 0) {
            return $this->json(['status' => 'error', 'message' => 'missing parameters'], 400);
        }

        $enabled = (bool) ($data['enable_ai_edit'] ?? false);

        $services = MediaWikiServices::getInstance();
        $user = $services->getUserFactory()->newFromId($userId);
        if (! $user || $user->isAnon()) {
            return $this->json(['status' => 'error', 'message' => "invalid user_id: {$userId}"], 400);
        }

        try {
            $userOptionsManager = $services->getUserOptionsManager();
            $userOptionsManager->setOption($user, 'ai-edit-as-mine', $enabled);
            $userOptionsManager->saveOptions($user);
        } catch (\Throwable $e) {
            return $this->json([
                'status' => 'error',
                'code' => 'exception',
                'message' => $e->getMessage(),
            ], 500);
        }

        return $this->json([
            'status' => 'success',
            'enabled' => $enabled,
            'user' => $user->getName(),
        ], 200);
    }
}
