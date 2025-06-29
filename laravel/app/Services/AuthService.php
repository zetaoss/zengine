<?php

namespace App\Services;

use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Redis;

class AuthService
{
    public static function me(): array|false
    {
        $userID = self::getValidUserID();

        return $userID ? self::getUserInfo($userID) : false;
    }

    private static function getValidUserID(): ?int
    {
        $prefix = config('app.wg_cookie_prefix');
        $userID = (int) request()->cookie("{$prefix}UserID");
        $userName = request()->cookie("{$prefix}UserName");

        if (! $userID || ! $userName) {
            return null;
        }

        if (self::isValidSession($prefix, $userID)) {
            return $userID;
        }

        if (self::isValidToken($prefix, $userID, $userName)) {
            return $userID;
        }

        return null;
    }

    private static function isValidSession(string $prefix, int $userID): bool
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

        return is_array($data) && ($data['data']['wsUserID'] ?? 0) == $userID;
    }

    private static function isValidToken(string $prefix, int $userID, string $userName): bool
    {
        $token = request()->cookie("{$prefix}Token");
        if (! $token) {
            return false;
        }

        $row = DB::connection('mwdb')->table('user')
            ->select('user_token')
            ->where('user_id', $userID)
            ->where('user_name', $userName)
            ->first();

        if (! $row || ! $row->user_token) {
            return false;
        }

        $expected = substr(hash_hmac('whirlpool', '1', $row->user_token, false), -32);

        return hash_equals($expected, $token);
    }

    private static function getUserInfo(int $userID): array
    {
        $cacheKey = 'userInfo:'.$userID;
        $cached = Cache::get($cacheKey);
        if ($cached !== null) {
            return $cached;
        }

        $groups = DB::connection('mwdb')->table('user_groups')
            ->where('ug_user', $userID)
            ->pluck('ug_group')
            ->toArray();

        $userInfo = [
            'avatar' => UserService::getUserAvatar($userID),
            'groups' => $groups,
        ];
        Cache::put($cacheKey, $userInfo, now()->addHours(1));

        return $userInfo;
    }
}
