<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Laravel\Socialite\Facades\Socialite;

class SocialController extends Controller
{
    private const SOCIALJOIN_TTL = 600;

    private const MWBRIDGE_TTL = 60;

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

        $row = DB::table('zetawiki.user_social')
            ->where('provider', $provider)
            ->where('social_id', $socialId)
            ->first();

        if (! $row) {
            try {
                DB::table('zetawiki.user_social')->insert([
                    'provider' => $provider,
                    'social_id' => $socialId,
                    'user_id' => null,
                ]);
            } catch (\Throwable $e) {
            }

            $row = DB::table('user_social')
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
        ], self::SOCIALJOIN_TTL);

        return redirect("/social-join/{$joinToken}");
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

    private function redis(): \Redis
    {
        $r = new \Redis;
        $r->connect((string) getenv('REDIS_HOST'));

        return $r;
    }
}
