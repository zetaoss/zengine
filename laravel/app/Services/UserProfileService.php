<?php

namespace App\Services;

use App\Models\UserProfile;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;

class UserProfileService
{
    private const CACHE_TTL = 3600;

    private static function cacheKey(int $userId): string
    {
        return "avatar:{$userId}";
    }

    private static function rowToProfileArray(object $row): array
    {
        return [
            'user_id' => (int) $row->user_id,
            'user_name' => (string) ($row->user_name ?? ''),
            't' => (int) ($row->t ?? 1),
            'ghash' => (string) ($row->ghash ?? ''),
        ];
    }

    private static function arrayToProfile(array $data): UserProfile
    {
        $profile = new UserProfile;
        $profile->forceFill([
            'user_id' => (int) ($data['user_id'] ?? 0),
            't' => (int) ($data['t'] ?? 1),
            'ghash' => (string) ($data['ghash'] ?? ''),
        ]);
        $profile->setAttribute('user_name', (string) ($data['user_name'] ?? ''));

        return $profile;
    }

    public static function toAvatarArray(?UserProfile $profile): ?array
    {
        if (! $profile) {
            return null;
        }

        return [
            'id' => (int) ($profile->user_id ?? 0),
            'name' => (string) ($profile->user_name ?? ''),
            't' => (int) ($profile->t ?? 1),
            'ghash' => (string) ($profile->ghash ?? ''),
        ];
    }

    public static function getUserProfile(int $userId): ?UserProfile
    {
        if ($userId < 1) {
            return null;
        }

        $profile = Cache::remember(
            self::cacheKey($userId),
            self::CACHE_TTL,
            static function () use ($userId) {
                $row = DB::table('zetawiki.user as A')
                    ->leftJoin('zetawiki.user_profiles as B', 'A.user_id', '=', 'B.user_id')
                    ->where('A.user_id', $userId)
                    ->first(['A.user_id', 'A.user_name', 'B.t', 'B.ghash']);

                return $row ? self::rowToProfileArray($row) : null;
            }
        );

        return is_array($profile) ? self::arrayToProfile($profile) : null;
    }

    public static function forget(int $userId): void
    {
        if ($userId < 1) {
            return;
        }

        Cache::forget(self::cacheKey($userId));
    }
}
