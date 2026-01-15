<?php

namespace App\Http\Controllers;

use App\Services\UserProfileService;

class UserProfileController extends Controller
{
    public function show(int $userId)
    {
        $userId = (int) $userId;
        if ($userId < 1) {
            return response()->json(['message' => 'Not Found'], 404);
        }

        $profile = UserProfileService::getUserProfile($userId);
        if (! $profile) {
            return response()->json(['message' => 'Not Found'], 404);
        }

        return response()->json([
            'name' => (string) ($profile->user_name ?? ''),
            't' => (int) ($profile->t ?? 1),
            'ghash' => (string) ($profile->ghash ?? ''),
        ]);
    }
}
