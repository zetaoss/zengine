<?php

namespace App\Http\Controllers;

use App\Models\UserSocial;
use Illuminate\Http\Request;

class SocialDeletionController extends Controller
{
    public function create(Request $request, string $provider)
    {
        if ($provider !== 'facebook') {
            abort(404);
        }

        $signedRequest = (string) $request->input('signed_request', '');
        $appSecret = (string) config('services.facebook.client_secret', '');
        $payload = $this->parseFacebookSignedRequest($signedRequest, $appSecret);

        if (! is_array($payload)) {
            return response()->json(['error' => 'invalid_signed_request'], 400);
        }

        $socialId = (string) ($payload['user_id'] ?? '');
        if ($socialId === '') {
            return response()->json(['error' => 'invalid_social_id'], 400);
        }

        $confirmationCode = bin2hex(random_bytes(16));

        $affected = UserSocial::query()
            ->where('provider', 'facebook')
            ->where('social_id', $socialId)
            ->update([
                'user_id' => null,
                'deletion_code' => $confirmationCode,
            ]);

        return response()->json([
            'url' => url("/auth/deletion/{$provider}/status/{$confirmationCode}"),
            'confirmation_code' => $confirmationCode,
            'deleted_links' => (int) $affected,
        ], 200, [], JSON_UNESCAPED_UNICODE);
    }

    public function status(string $provider, string $code)
    {
        if ($provider !== 'facebook') {
            abort(404);
        }

        if (! preg_match('/^[a-zA-Z0-9_-]{1,100}$/', $code)) {
            abort(404);
        }

        $row = UserSocial::query()
            ->where('provider', 'facebook')
            ->where('deletion_code', $code)
            ->first();
        if (! $row) {
            abort(404);
        }

        return response()->json([
            'status' => 'completed',
            'confirmation_code' => $code,
            'deleted_links' => 1,
        ]);
    }

    private function parseFacebookSignedRequest(string $signedRequest, string $appSecret): ?array
    {
        if ($signedRequest === '' || $appSecret === '' || ! str_contains($signedRequest, '.')) {
            return null;
        }

        [$encodedSig, $encodedPayload] = explode('.', $signedRequest, 2);
        $sig = $this->base64UrlDecode($encodedSig);
        $payloadRaw = $this->base64UrlDecode($encodedPayload);

        if (! is_string($sig) || ! is_string($payloadRaw) || $sig === '' || $payloadRaw === '') {
            return null;
        }

        $payload = json_decode($payloadRaw, true);
        if (! is_array($payload)) {
            return null;
        }

        $alg = strtoupper((string) ($payload['algorithm'] ?? ''));
        if ($alg !== 'HMAC-SHA256') {
            return null;
        }

        $expected = hash_hmac('sha256', $encodedPayload, $appSecret, true);
        if (! hash_equals($expected, $sig)) {
            return null;
        }

        return $payload;
    }

    private function base64UrlDecode(string $value): ?string
    {
        $value = strtr($value, '-_', '+/');
        $padding = strlen($value) % 4;
        if ($padding > 0) {
            $value .= str_repeat('=', 4 - $padding);
        }

        $decoded = base64_decode($value, true);

        return is_string($decoded) ? $decoded : null;
    }
}
