<?php

namespace App\Services;

use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Redis;

class AuthService
{
    public static function me(): ?array
    {
        $userId = self::getValidUserId();

        return $userId ? self::getUserInfo($userId) : null;
    }

    private static function getValidUserId(): ?int
    {
        $prefix = config('app.wg_cookie_prefix');
        $userId = (int) request()->cookie("{$prefix}UserID");
        $userName = request()->cookie("{$prefix}UserName");

        if (! $userId || ! $userName) {
            return null;
        }

        if (self::isValidSession($prefix, $userId)) {
            return $userId;
        }

        if (self::isValidToken($prefix, $userId, $userName)) {
            return $userId;
        }

        return null;
    }

    private static function isValidSession(string $prefix, int $userId): bool
    {
        $sessionKey = request()->cookie("{$prefix}_session");
        if (! $sessionKey) {
            return false;
        }

        $sessionData = Redis::connection('mwsession')->get('zetawiki:MWSession:'.$sessionKey);
        if (! $sessionData) {
            return false;
        }

        $data = @unserialize($sessionData, ['allowed_classes' => false]);

        return is_array($data) && ($data['data']['wsUserID'] ?? 0) == $userId;
    }

    private static function isValidToken(string $prefix, int $userId, string $userName): bool
    {
        $token = request()->cookie("{$prefix}Token");
        if (! $token) {
            return false;
        }

        $row = DB::connection('mwdb')->table('user')
            ->select('user_token')
            ->where('user_id', $userId)
            ->where('user_name', $userName)
            ->first();

        if (! $row || ! $row->user_token) {
            return false;
        }

        $expected = substr(hash_hmac('whirlpool', '1', $row->user_token, false), -32);

        return hash_equals($expected, $token);
    }

    private static function getUserInfo(int $userId): array
    {
        $key = "userInfo:$userId";
        if ($cached = Cache::get($key)) {
            return $cached;
        }

        $groups = DB::connection('mwdb')->table('user_groups')
            ->where('ug_user', $userId)
            ->pluck('ug_group')
            ->toArray();

        $userInfo = [
            'avatar' => AvatarService::getAvatarById($userId),
            'groups' => $groups,
        ];
        Cache::put($key, $userInfo, 3600);

        return $userInfo;
    }

    public static function forgetUserInfo(int $userId): void
    {
        Cache::forget("userInfo:$userId");

        AvatarService::forgetAvatar($userId);
    }
}
