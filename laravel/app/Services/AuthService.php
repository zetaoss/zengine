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
        $prefix = getenv('WG_COOKIE_PREFIX');
        $userID = (int) ($_COOKIE["{$prefix}UserID"] ?? 0);
        $userName = $_COOKIE["{$prefix}UserName"] ?? null;

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
        $sessionKey = $_COOKIE["{$prefix}_session"] ?? null;
        if (! $sessionKey) {
            return false;
        }

        $sessionData = Redis::connection('mwsession')->get('zetawiki:MWSession:'.$sessionKey);
        if (! $sessionData) {
            return false;
        }

        $data = @unserialize($sessionData);

        return is_array($data) && ($data['data']['wsUserID'] ?? 0) == $userID;
    }

    private static function isValidToken(string $prefix, int $userID, string $userName): bool
    {
        $token = $_COOKIE["{$prefix}Token"] ?? null;
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
        Cache::put($cacheKey, $userInfo);

        return $userInfo;
    }
}
