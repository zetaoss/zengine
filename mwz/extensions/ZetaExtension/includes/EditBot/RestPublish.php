<?php

namespace ZetaExtension\EditBot;

use ContentHandler;
use MediaWiki\MediaWikiServices;
use MediaWiki\Rest\SimpleHandler;
use Title;
use User;

class RestPublish extends SimpleHandler
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
        $titleText = (string) ($data['title'] ?? '');
        $content = (string) ($data['text'] ?? '');
        $summary = (string) ($data['summary'] ?? '');
        $requestType = (string) ($data['request_type'] ?? '');

        if ($userId <= 0 || $titleText === '') {
            return $this->json(['status' => 'error', 'message' => 'missing parameters'], 400);
        }

        $services = MediaWikiServices::getInstance();
        $user = User::newFromId($userId);
        if (! $user || $user->isAnon()) {
            return $this->json(['status' => 'error', 'message' => "invalid user_id: {$userId}"], 400);
        }

        $title = Title::newFromText($titleText);
        if (! $title) {
            return $this->json(['status' => 'error', 'message' => 'invalid title'], 400);
        }

        try {
            $wikiPage = $services->getWikiPageFactory()->newFromTitle($title);
            $contentObj = ContentHandler::makeContent($content, $title);

            $flags = EDIT_INTERNAL;
            if ($requestType === 'create') {
                $flags |= EDIT_NEW;
            } else {
                $flags |= EDIT_UPDATE;
            }

            // Perform edit as the specified user
            $status = $wikiPage->doUserEditContent(
                $contentObj,
                $user,
                $summary,
                $flags
            );

            if (! $status->isOK()) {
                $errors = $status->getErrorsArray();
                $code = 'editfailed';
                if (! empty($errors) && isset($errors[0][0])) {
                    $code = $errors[0][0];
                }

                return $this->json([
                    'status' => 'error',
                    'code' => $code,
                    'message' => 'edit failed',
                    'errors' => $errors,
                ], 500);
            }

            $value = $status->getValue();
            $revid = 0;
            if (isset($value['revision-record'])) {
                $revid = $value['revision-record']->getId();
            } elseif (isset($value['revision'])) {
                $revid = $value['revision']->getId();
            }

            return $this->json([
                'status' => 'success',
                'revid' => $revid,
                'title' => $title->getPrefixedText(),
                'user' => $user->getName(),
            ], 200);
        } catch (\Throwable $e) {
            return $this->json([
                'status' => 'error',
                'message' => $e->getMessage(),
            ], 500);
        }
    }
}
