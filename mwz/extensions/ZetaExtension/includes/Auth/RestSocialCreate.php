<?php

namespace ZetaExtension\Auth;

use MediaWiki\MediaWikiServices;
use MediaWiki\Rest\SimpleHandler;
use User;
use Wikimedia\Rdbms\IDatabase;

class RestSocialCreate extends SimpleHandler
{
    private const SOCIALJOIN_TTL = 600;

    private const MWBRIDGE_TTL = 60;

    private const MAX_USERNAME_LEN = 80;

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

    private function normalizeInputUsername(string $username): string
    {
        $username = preg_replace('/ +/u', ' ', $username);

        return trim((string) $username);
    }

    private function redis(): \Redis
    {
        $r = new \Redis;
        $r->connect(getenv('REDIS_HOST'));

        return $r;
    }

    private function putToken(string $prefix, array $payload, int $ttlSeconds): string
    {
        $token = bin2hex(random_bytes(32));
        $key = "{$prefix}:{$token}";
        $this->redis()->setex($key, $ttlSeconds, json_encode($payload, JSON_UNESCAPED_UNICODE));

        return $token;
    }

    private function getToken(string $prefix, string $token): ?array
    {
        if ($token === '') {
            return null;
        }
        $key = "{$prefix}:{$token}";
        $raw = $this->redis()->get($key);
        if (! is_string($raw) || $raw === '') {
            return null;
        }
        $data = json_decode($raw, true);

        return is_array($data) ? $data : null;
    }

    private function popToken(string $prefix, string $token): ?array
    {
        if ($token === '') {
            return null;
        }
        $key = "{$prefix}:{$token}";
        $raw = $this->redis()->getDel($key);
        if (! is_string($raw) || $raw === '') {
            return null;
        }
        $data = json_decode($raw, true);

        return is_array($data) ? $data : null;
    }

    private function dryRunUsername(string $username): array
    {
        $username = $this->normalizeInputUsername($username);

        if ($username === '' || mb_strlen($username) > self::MAX_USERNAME_LEN) {
            return [
                'status' => 'success',
                'dryrun' => true,
                'can_create' => false,
                'input' => $username,
                'name' => $username,
                'normalized' => false,
                'errors' => [['code' => 'invalid_username']],
                'messages' => ['invalid username'],
            ];
        }

        $services = MediaWikiServices::getInstance();
        $userNameUtils = $services->getUserNameUtils();
        $authManager = $services->getAuthManager();

        $canonical = $userNameUtils->getCanonical($username);
        $finalName = (is_string($canonical) && $canonical !== '') ? $canonical : $username;
        $normalized = $finalName !== $username;

        if (! $userNameUtils->isCreatable($finalName)) {
            return [
                'status' => 'success',
                'dryrun' => true,
                'can_create' => false,
                'input' => $username,
                'name' => $finalName,
                'normalized' => $normalized,
                'errors' => [['code' => 'username_not_allowed']],
                'messages' => ['username not allowed'],
            ];
        }

        if ($authManager->userExists($finalName)) {
            return [
                'status' => 'success',
                'dryrun' => true,
                'can_create' => false,
                'input' => $username,
                'name' => $finalName,
                'normalized' => $normalized,
                'errors' => [['code' => 'userexists']],
                'messages' => ['username already exists'],
            ];
        }

        $messages = [];
        if ($normalized) {
            $messages[] = "당신의 사용자 이름은 기술적 한계로 인해 {$finalName}(으)로 조정됩니다.";
        }

        return [
            'status' => 'success',
            'dryrun' => true,
            'can_create' => true,
            'input' => $username,
            'name' => $finalName,
            'normalized' => $normalized,
            'messages' => $messages,
        ];
    }

    private function mwBridgeUrl(string $token): string
    {
        return '/w/rest.php/social/bridge?token='.rawurlencode($token);
    }

    private function updateUserSocial(IDatabase $dbw, int $userSocialId, int $userId): void
    {
        $dbw->update(
            'user_social',
            ['user_id' => $userId],
            ['id' => $userSocialId],
            __METHOD__
        );
    }

    private function createUserDirect(string $finalName, ?string $email, ?string $realname): ?User
    {
        $u = User::newFromName($finalName, 'creatable');
        if (! $u || ! $u->isAnon()) {
            return null;
        }

        if (is_string($email) && $email !== '') {
            $u->setEmail($email);
            $u->setEmailAuthenticationTimestamp(null);
        }

        if (is_string($realname) && $realname !== '') {
            $u->setRealName($realname);
        }

        $u->addToDatabase();
        if ((int) $u->getId() < 1) {
            return null;
        }

        return $u;
    }

    public function execute()
    {
        $data = $this->readJsonBody();
        if (! $data) {
            return $this->json(['status' => 'error', 'message' => 'invalid json body'], 400);
        }

        $token = (string) ($data['token'] ?? '');
        $dryrun = (bool) ($data['dryrun'] ?? false);

        $payload = $dryrun ? $this->getToken('socialjoin', $token) : $this->popToken('socialjoin', $token);
        if (! $payload) {
            return $this->json(['status' => 'error', 'code' => 'invalid_token', 'message' => 'invalid token'], 403);
        }

        $username = $this->normalizeInputUsername((string) ($data['username'] ?? ''));
        if ($username === '' || mb_strlen($username) > self::MAX_USERNAME_LEN) {
            if (! $dryrun) {
                $newToken = $this->putToken('socialjoin', $payload, self::SOCIALJOIN_TTL);

                return $this->json(['status' => 'error', 'message' => 'invalid username', 'token' => $newToken], 422);
            }

            return $this->json($this->dryRunUsername($username), 200);
        }

        if ($dryrun) {
            return $this->json($this->dryRunUsername($username), 200);
        }

        $dry = $this->dryRunUsername($username);
        if (! (($dry['status'] ?? '') === 'success' && ($dry['can_create'] ?? false) === true)) {
            $newToken = $this->putToken('socialjoin', $payload, self::SOCIALJOIN_TTL);

            return $this->json([
                'status' => 'error',
                'message' => 'username not allowed',
                'token' => $newToken,
                'debug' => ['dryrun' => $dry],
            ], 422);
        }

        $finalName = (string) ($dry['name'] ?? $username);

        $email = isset($payload['email']) ? trim((string) $payload['email']) : null;
        $realname = isset($payload['realname']) ? trim((string) $payload['realname']) : null;

        $services = MediaWikiServices::getInstance();
        $lbFactory = $services->getDBLoadBalancerFactory();
        $dbw = $lbFactory->getMainLB()->getConnection(DB_PRIMARY);

        $userSocialId = (int) ($payload['user_social_id'] ?? 0);

        $dbw->startAtomic(__METHOD__);
        try {
            $user = $this->createUserDirect($finalName, $email, $realname);
            if (! $user) {
                throw new \RuntimeException('failed to create user');
            }

            $userId = (int) $user->getId();
            if ($userSocialId > 0) {
                $this->updateUserSocial($dbw, $userSocialId, $userId);
            }

            $dbw->endAtomic(__METHOD__);
        } catch (\Throwable $e) {
            $dbw->cancelAtomic(__METHOD__);
            $newToken = $this->putToken('socialjoin', $payload, self::SOCIALJOIN_TTL);

            return $this->json([
                'status' => 'error',
                'message' => $e->getMessage(),
                'token' => $newToken,
            ], 400);
        }

        $bridgeToken = $this->putToken('mwbridge', [
            'user_id' => (int) $user->getId(),
            'returnto' => (string) ($payload['returnto'] ?? ''),
        ], self::MWBRIDGE_TTL);

        return $this->json([
            'status' => 'success',
            'user_id' => (int) $user->getId(),
            'username' => $user->getName(),
            'redirect' => $this->mwBridgeUrl($bridgeToken),
        ], 200);
    }
}
