<?php

namespace App\Http\Controllers;

use App\Models\Avatar;
use App\Services\AvatarService;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\Facades\Http;

class AuthController extends Controller
{
    public function me()
    {
        $id = auth()->id();
        if (! $id) {
            return response()->json(['me' => null]);
        }

        $me = auth()->user()->toArray();
        $me['avatar'] = AvatarService::getAvatarById((int) $id);

        return response()->json(['me' => $me]);
    }

    public function verifyGravatar(Request $request)
    {
        Gate::authorize('unblocked');

        $data = $request->validate([
            'ghash' => ['required', 'string', 'size:32', 'regex:/^[0-9a-f]{32}$/'],
        ]);

        $hash = strtolower($data['ghash']);

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

        $avatar = Avatar::firstOrNew(['user_id' => $userId]);
        $avatar->t = (int) $data['t'];

        if (array_key_exists('ghash', $data)) {
            $hash = $data['ghash'];

            if ($hash === null || trim($hash) === '') {
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

        AvatarService::forgetAvatar($userId);

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
