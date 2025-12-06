<?php

namespace App\Services;

use App\Models\User;
use Illuminate\Support\Facades\Cache;

class AvatarService
{
    public static function getAvatarById(int $userId): ?array
    {
        $key = "avatar:$userId";

        return Cache::remember($key, 3600, function () use ($userId) {
            $user = User::with('avatarRelation')->find($userId);

            if (! $user) {
                return null;
            }

            return $user->avatar;
        });
    }
}
