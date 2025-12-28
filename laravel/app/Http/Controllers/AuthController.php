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
            ->first(['ghint']);

        return response()->json([
            'ghint' => $row?->ghint ?? '',
        ]);
    }

    public function verifyGravatar(Request $request)
    {
        Gate::authorize('unblocked');

        $data = $request->validate([
            'ghash' => ['required', 'string', 'size:32', 'regex:/^[0-9a-f]{32}$/'],
        ]);

        $ghash = strtolower((string) $data['ghash']);
        if ($ghash === '') {
            return response()->json(['ok' => false], 422);
        }

        if (! $this->gravatarExists($ghash)) {
            return response()->json(['ok' => false], 404);
        }

        return response()->json([
            'ok' => true,
            'ghash' => $ghash,
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
            'ghint' => ['sometimes', 'nullable', 'string', 'max:255'],
        ]);

        $t = (int) $data['t'];

        $avatar = Avatar::firstOrNew(['user_id' => $userId]);
        $avatar->t = $t;

        if ($t === 3) {
            $reqHash = strtolower((string) ($data['ghash'] ?? ''));
            $reqHint = trim((string) ($data['ghint'] ?? ''));

            if ($reqHash !== '') {
                if ($reqHint === '') {
                    return response()->json(['message' => 'ghint required for gravatar'], 422);
                }

                if (! $this->gravatarExists($reqHash)) {
                    return response()->json(['message' => 'Gravatar not found'], 400);
                }

                $avatar->ghash = $reqHash;
                $avatar->ghint = $reqHint;
            } else {
                $existingHash = trim((string) ($avatar->ghash ?? ''));
                if ($existingHash === '') {
                    return response()->json(['message' => 'gravatar not configured'], 422);
                }
            }
        }

        $avatar->save();

        AvatarService::forget($userId);

        return response()->json([
            'avatar' => AvatarService::getAvatarById($userId),
        ]);
    }

    private function gravatarExists(string $ghash): bool
    {
        $url = "https://www.gravatar.com/avatar/{$ghash}?d=404";

        try {
            $resp = Http::timeout(3)->head($url);

            return $resp->status() === 200;
        } catch (\Throwable) {
            return false;
        }
    }
}
