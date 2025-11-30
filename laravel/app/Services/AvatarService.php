<?php

namespace App\Services;

use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;

class AvatarService
{
    private const CACHE_TTL = 3600;

    private static function zeroAvatar(): array
    {
        return [
            'id' => 0,
            'name' => '?',
            't' => 2,
            'ghash' => '',
        ];
    }

    public static function getAvatarById(int $userId): ?array
    {
        if ($userId === 0) {
            return self::zeroAvatar();
        }

        $key = "avatar:id:$userId";
        if ($cached = Cache::get($key)) {
            return $cached;
        }

        $mwdb = DB::connection('mwdb')->getDatabaseName();
        $row = DB::selectOne("
            SELECT A.user_name, B.t, B.ghash
            FROM {$mwdb}.user A
            LEFT JOIN profiles B ON A.user_id = B.user_id
            WHERE A.user_id = ?
        ", [$userId]);

        if (! $row) {
            return null;
        }

        $avatar = [
            'id' => $userId,
            'name' => $row->user_name,
            't' => $row->t,
            'ghash' => $row->ghash,
        ];

        Cache::put($key, $avatar, self::CACHE_TTL);

        return $avatar;
    }

    public static function getAvatarByName(string $username): ?array
    {
        $username = trim($username);

        if ($username === '' || $username === '?') {
            return self::zeroAvatar();
        }

        $key = "avatar:name:$username";
        if ($cached = Cache::get($key)) {
            return $cached;
        }

        $mwdb = DB::connection('mwdb')->getDatabaseName();

        $row = DB::selectOne("
            SELECT A.user_id, A.user_name, B.t, B.ghash
            FROM {$mwdb}.user A
            LEFT JOIN profiles B ON A.user_id = B.user_id
            WHERE A.user_name = ?
        ", [$username]);

        if (! $row) {
            return null;
        }

        $userId = (int) $row->user_id;

        $avatar = [
            'id' => $userId,
            'name' => $row->user_name,
            't' => $row->t,
            'ghash' => $row->ghash,
        ];

        Cache::put($key, $avatar, self::CACHE_TTL);

        return $avatar;
    }
}
