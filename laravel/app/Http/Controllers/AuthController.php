<?php

namespace App\Http\Controllers;

use App\Models\Avatar;
use App\Services\AvatarService;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\Facades\Http;

class AuthController extends Controller
{
    public function me(Request $request)
    {
        $user = auth()->user();
        if (! $user) {
            return response()->json(['me' => null]);
        }

        $id = (int) auth()->id();
        if ($id < 1) {
            return response()->json(['me' => null]);
        }

        $me = (array) $user->toArray();
        $me['avatar'] = AvatarService::getAvatarById($id);

        return response()->json(['me' => $me]);
    }

    public function getGravatar()
    {
        $userId = (int) auth()->id();
        if ($userId < 1) {
            return response()->json(['message' => 'Unauthenticated'], 401);
        }

        $row = Avatar::query()
            ->where('user_id', $userId)
            ->first(['gravatar']);

        return response()->json([
            'gravatar' => $row?->gravatar ?? '',
        ]);
    }

    public function verifyGravatar(Request $request)
    {
        Gate::authorize('unblocked');

        $data = $request->validate([
            'email' => ['required', 'email:rfc', 'max:255'],
        ]);

        $email = trim((string) $data['email']);
        if ($email === '') {
            return response()->json(['ok' => false], 422);
        }

        $hash = md5(strtolower(trim($email)));

        if (! $this->gravatarExists($hash)) {
            return response()->json(['ok' => false], 404);
        }

        return response()->json([
            'ok' => true,
            'ghash' => $hash,
        ]);
    }

    public function updateAvatar(Request $request)
    {
        Gate::authorize('unblocked');

        $userId = (int) auth()->id();
        if ($userId < 1) {
            return response()->json(['message' => 'Unauthenticated'], 401);
        }

        $data = $request->validate([
            't' => ['required', 'integer', 'min:1', 'max:3'],
            'ghash' => ['sometimes', 'nullable', 'string', 'size:32', 'regex:/^[0-9a-f]{32}$/'],
        ]);

        $t = (int) $data['t'];

        $avatar = Avatar::firstOrNew(['user_id' => $userId]);
        $avatar->t = $t;

        if ($t === 3 && ! array_key_exists('ghash', $data)) {
            return response()->json(['message' => 'ghash required for gravatar'], 422);
        }

        if (array_key_exists('ghash', $data)) {
            $hash = $data['ghash'];

            if ($hash === null || trim($hash) === '') {
                if ($t === 3) {
                    return response()->json(['message' => 'ghash required for gravatar'], 422);
                }
                $avatar->ghash = null;
            } else {
                $hash = strtolower($hash);

                if (! $this->gravatarExists($hash)) {
                    return response()->json(['message' => 'Gravatar not found'], 400);
                }

                $avatar->ghash = $hash;
            }
        }

        $avatar->save();

        AvatarService::forget($userId);

        return response()->json([
            'avatar' => AvatarService::getAvatarById($userId),
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
