<?php

namespace App\Http\Controllers;

use App\Models\Avatar;
use App\Services\AuthService;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Http;

class AuthController extends Controller
{
    public function logout()
    {
        $prefix = getenv('WG_COOKIE_PREFIX');
        setcookie($prefix.'Token', '', time() - 3600, '/');
        setcookie($prefix.'UserID', '', time() - 3600, '/');
        setcookie($prefix.'UserName', '', time() - 3600, '/');
        setcookie($prefix.'_session', '', time() - 3600, '/');

        return response()->noContent();
    }

    public function me(Request $request)
    {
        $me = $request->attributes->get('me');

        return $me ?? (object) [];
    }

    public function verifyGravatar(Request $request)
    {
        $request->validate([
            'email' => ['required', 'email'],
        ]);

        $email = strtolower(trim((string) $request->query('email')));
        $hash = md5($email);

        if (! $this->gravatarExists($hash)) {
            return response()->json(['ok' => 0], 404);
        }

        return response()->json([
            'ok' => true,
            'ghash' => $hash,
        ]);
    }

    public function updateAvatar(Request $request)
    {
        $me = $request->attributes->get('me');
        $userId = (int) (($me['avatar']['id'] ?? 0));

        if ($userId < 1) {
            return response()->json(['message' => 'Unauthenticated'], 401);
        }

        $data = $request->validate([
            't' => ['required', 'integer', 'min:1', 'max:3'],
            'gravatar' => ['sometimes', 'nullable', 'string'],
        ]);

        $t = (int) $data['t'];

        $avatar = Avatar::firstOrNew(['user_id' => $userId]);
        $avatar->t = $t;

        if (array_key_exists('gravatar', $data)) {
            $g = $data['gravatar'];

            if ($g === null || trim((string) $g) === '') {
                $avatar->gravatar = null;
                $avatar->ghash = null;
            } else {
                $request->validate([
                    'gravatar' => ['required', 'email'],
                ]);

                $email = strtolower(trim((string) $g));
                $hash = md5($email);

                if (! $this->gravatarExists($hash)) {
                    return response()->json(['message' => 'Gravatar not found'], 400);
                }

                $avatar->gravatar = $email;
                $avatar->ghash = $hash;
            }
        }

        $avatar->save();

        AuthService::forgetUserInfo($userId);

        return response()->json([
            'avatar' => [
                't' => (int) ($avatar->t ?? 1),
                'gravatar' => (string) ($avatar->gravatar ?? ''),
                'ghash' => (string) ($avatar->ghash ?? ''),
            ],
        ]);
    }

    private function gravatarExists(string $hash): bool
    {
        $url = "https://www.gravatar.com/avatar/{$hash}?d=404";

        try {
            $resp = Http::timeout(3)->head($url);

            return $resp->status() === 200;
        } catch (\Throwable) {
            return false;
        }
    }
}
