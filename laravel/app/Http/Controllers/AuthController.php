<?php

namespace App\Http\Controllers;

use App\Models\UserProfile;
use App\Services\UserProfileService;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;

class AuthController extends Controller
{
    public function me(Request $request)
    {
        $user = auth()->user();
        if (! $user) {
            return response()->json(['me' => null]);
        }

        $u = (array) $user->toArray();

        $id = (int) ($u['id'] ?? 0);
        if ($id < 1) {
            return response()->json(['me' => null]);
        }

        $name = (string) ($u['name'] ?? '');
        $groups = $u['groups'] ?? [];
        if (! is_array($groups)) {
            $groups = [];
        }

        return response()->json([
            'me' => [
                'id' => $id,
                'name' => $name,
                'groups' => $groups,
            ],
        ]);
    }

    public function getAvatar()
    {
        $userId = (int) auth()->id();
        if ($userId < 1) {
            return response()->json(['message' => 'Unauthenticated'], 401);
        }

        $row = UserProfile::query()
            ->where('user_id', $userId)
            ->first(['t', 'ghint']);

        return response()->json([
            't' => (int) ($row?->t ?? 1),
            'ghint' => (string) ($row?->ghint ?? ''),
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

        $avatar = UserProfile::firstOrNew(['user_id' => $userId]);
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

        UserProfileService::forget($userId);
        $this->purgeAvatarQuietly($userId);

        return response()->json([
            'avatar' => UserProfileService::toAvatarArray(
                UserProfileService::getUserProfile($userId)
            ),
        ]);
    }

    private function purgeAvatarQuietly(int $userId): void
    {
        if ($userId < 1) {
            return;
        }

        $internalApi = config('services.zavatar.internal_api');
        $internalKey = config('services.zavatar.internal_key');

        if (! $internalApi || ! $internalKey) {
            Log::error('Zavatar config missing');

            return;
        }

        try {
            $url = "${internalApi}/internal/purge/u/${userId}";
            $response = Http::timeout(2)
                ->withHeaders(['X-Internal-Key' => $internalKey])
                ->post($url);

            if ($response->failed()) {
                Log::error('Zavatar purge failed', [
                    'url' => $url,
                    'user_id' => $userId,
                    'status' => $response->status(),
                    'body' => $response->body(),
                ]);
            } else {
                Log::info('Zavatar purge success', ['user_id' => $userId]);
            }

        } catch (\Throwable $e) {
            Log::error('Zavatar purge exception', [
                'user_id' => $userId,
                'message' => $e->getMessage(),
            ]);
        }
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
