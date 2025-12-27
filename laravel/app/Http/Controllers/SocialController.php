<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Http;
use Laravel\Socialite\Facades\Socialite;

class SocialController extends Controller
{
    private const SOCIALJOIN_TTL = 600;

    private const MWBRIDGE_TTL = 60;

    private const MAX_USERNAME_LEN = 80;

    public function redirect(Request $request, string $provider)
    {
        $returnto = (string) $request->input('returnto', '');
        $returnto = $this->sanitizeReturnto($returnto);

        $request->session()->put('returnto', $returnto);

        return Socialite::driver($provider)->redirect();
    }

    public function callback(Request $request, string $provider)
    {
        try {
            $socialUser = Socialite::driver($provider)->user();
        } catch (\Throwable $e) {
            return redirect('/login?error=social_auth_failed');
        }

        $socialId = (string) $socialUser->getId();
        if ($socialId === '') {
            return redirect('/login?error=invalid_social_id');
        }

        $returnto = (string) $request->session()->get('returnto', '');
        $request->session()->forget('returnto');

        $mwdb = DB::connection('mwdb');

        $row = $mwdb->table('user_social')
            ->where('provider', $provider)
            ->where('social_id', $socialId)
            ->first();

        if (! $row) {
            try {
                $mwdb->table('user_social')->insert([
                    'provider' => $provider,
                    'social_id' => $socialId,
                    'user_id' => null,
                ]);
            } catch (\Throwable $e) {
            }

            $row = $mwdb->table('user_social')
                ->where('provider', $provider)
                ->where('social_id', $socialId)
                ->first();
        }

        if (! $row) {
            return redirect('/login?error=social_link_failed');
        }

        $userId = (int) ($row->user_id ?? 0);

        if ($userId > 0) {
            $bridgeToken = $this->putToken('mwbridge', [
                'user_id' => $userId,
                'returnto' => $returnto,
            ], self::MWBRIDGE_TTL);

            return redirect($this->mwBridgeUrl($bridgeToken));
        }

        $joinToken = $this->putToken('socialjoin', [
            'user_social_id' => (int) $row->id,
            'provider' => $provider,
            'social_id' => $socialId,
            'returnto' => $returnto,
            'email' => (string) ($socialUser->getEmail() ?? ''),
            'realname' => trim((string) ($socialUser->getName() ?? '')),
        ], self::SOCIALJOIN_TTL);

        return redirect("/social-join/{$joinToken}");
    }

    public function join(Request $request)
    {
        $token = (string) $request->input('token', '');
        $username = trim((string) $request->input('username', ''));

        $payload = $this->popToken('socialjoin', $token);
        if (! $payload) {
            return response()->json(['status' => 'error', 'message' => 'invalid token'], 403);
        }

        if ($username === '' || mb_strlen($username) > self::MAX_USERNAME_LEN) {
            return $this->errorWithNewToken($payload, 422, 'invalid username');
        }

        $dry = $this->mwDryRunUsername($username);
        if (! (($dry['status'] ?? '') === 'success' && ($dry['can_create'] ?? false) === true)) {
            return $this->errorWithNewToken($payload, 422, 'username not allowed', ['mw_dryrun' => $dry]);
        }

        $finalName = (string) ($dry['name'] ?? $username);

        $mw = $this->createMwAccount($finalName, $payload);

        if (($mw['user_id'] ?? 0) < 1) {
            return $this->errorWithNewToken($payload, 500, 'failed to create account', $mw);
        }

        $userId = (int) $mw['user_id'];

        DB::connection('mwdb')->table('user_social')
            ->where('id', (int) ($payload['user_social_id'] ?? 0))
            ->update(['user_id' => $userId]);

        $bridgeToken = $this->putToken('mwbridge', [
            'user_id' => $userId,
            'returnto' => (string) ($payload['returnto'] ?? ''),
        ], self::MWBRIDGE_TTL);

        return response()->json([
            'status' => 'success',
            'redirect' => $this->mwBridgeUrl($bridgeToken),
        ]);
    }

    private function errorWithNewToken(array $payload, int $httpCode, string $message, array $extra = [])
    {
        $newToken = $this->putToken('socialjoin', [
            'user_social_id' => (int) ($payload['user_social_id'] ?? 0),
            'provider' => (string) ($payload['provider'] ?? ''),
            'social_id' => (string) ($payload['social_id'] ?? ''),
            'returnto' => (string) ($payload['returnto'] ?? ''),
            'email' => (string) ($payload['email'] ?? ''),
            'realname' => (string) ($payload['realname'] ?? ''),
        ], self::SOCIALJOIN_TTL);

        $res = [
            'status' => 'error',
            'message' => $message,
            'token' => $newToken,
        ];

        if (! empty($extra)) {
            $res['debug'] = $extra;
        }

        return response()->json($res, $httpCode);
    }

    private function sanitizeReturnto(string $returnto): string
    {
        $returnto = trim($returnto);

        if ($returnto === '' || strlen($returnto) > 200) {
            return '';
        }

        if (preg_match('~://|\\\\|\.\.~', $returnto)) {
            return '';
        }

        if (! preg_match('~^[\pL\pN _:\-/().%]+$~u', $returnto)) {
            return '';
        }

        return $returnto;
    }

    private function mwBridgeUrl(string $token): string
    {
        return '/w/rest.php/social/bridge?token='.rawurlencode($token);
    }

    private function mwBaseUrl(): string
    {
        $base = (string) config('app.url');

        return rtrim($base, '/');
    }

    private function putToken(string $prefix, array $payload, int $ttlSeconds): string
    {
        $token = bin2hex(random_bytes(32));
        $key = "{$prefix}:{$token}";

        $this->redis()->setex($key, $ttlSeconds, json_encode($payload, JSON_UNESCAPED_UNICODE));

        return $token;
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

    private function mwDryRunUsername(string $username): array
    {
        $url = $this->mwBaseUrl().'/w/rest.php/social-create';

        try {
            $resp = Http::timeout(10)
                ->acceptJson()
                ->asJson()
                ->post($url, [
                    'username' => $username,
                    'dryrun' => true,
                ]);
        } catch (\Throwable $e) {
            return [
                'status' => 'error',
                'message' => $e->getMessage(),
            ];
        }

        if (! $resp->ok()) {
            return [
                'status' => 'error',
                'mw_status' => $resp->status(),
                'mw_body' => $resp->body(),
            ];
        }

        $json = null;
        try {
            $json = $resp->json();
        } catch (\Throwable $e) {
            $json = null;
        }

        return is_array($json) ? $json : [
            'status' => 'error',
            'message' => 'invalid mw response',
        ];
    }

    private function createMwAccount(string $username, array $payload): array
    {
        $url = $this->mwBaseUrl().'/w/rest.php/social-create';

        try {
            $resp = Http::timeout(10)
                ->acceptJson()
                ->asJson()
                ->post($url, [
                    'username' => $username,
                    'email' => (string) ($payload['email'] ?? ''),
                    'realname' => (string) ($payload['realname'] ?? ''),
                    'provider' => (string) ($payload['provider'] ?? ''),
                    'social_id' => (string) ($payload['social_id'] ?? ''),
                ]);
        } catch (\Throwable $e) {
            return [
                'user_id' => 0,
                'mw_status' => 0,
                'mw_body' => null,
                'exception' => $e->getMessage(),
            ];
        }

        $raw = $resp->body();
        $json = null;

        try {
            $json = $resp->json();
        } catch (\Throwable $e) {
            $json = null;
        }

        $userId = 0;
        if (is_array($json) && ($json['status'] ?? '') === 'success') {
            $userId = (int) ($json['user_id'] ?? 0);
        }

        return [
            'user_id' => $userId,
            'mw_status' => $resp->status(),
            'mw_body' => $json ?? $raw,
        ];
    }

    private function redis(): \Redis
    {
        $r = new \Redis;
        $r->connect((string) getenv('REDIS_HOST'));

        return $r;
    }
}
