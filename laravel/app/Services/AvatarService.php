<?php

namespace App\Services;

use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;

class AvatarService
{
    private const CACHE_TTL = 3600;

    private static function cacheKey(int $userId): string
    {
        return "avatar:{$userId}";
    }

    private static function rowToAvatar(object $row): array
    {
        return [
            'id' => (int) $row->user_id,
            'name' => (string) ($row->user_name ?? ''),
            't' => (int) ($row->t ?? 1),
            'ghash' => (string) ($row->ghash ?? ''),
        ];
    }

    public static function getAvatarById(int $userId): ?array
    {
        if ($userId < 1) {
            return null;
        }

        $avatar = Cache::remember(
            self::cacheKey($userId),
            self::CACHE_TTL,
            static function () use ($userId) {
                $row = DB::table('zetawiki.user as A')
                    ->leftJoin('profiles as B', 'A.user_id', '=', 'B.user_id')
                    ->where('A.user_id', $userId)
                    ->first(['A.user_id', 'A.user_name', 'B.t', 'B.ghash']);

                return $row ? self::rowToAvatar($row) : null;
            }
        );

        return is_array($avatar) ? $avatar : null;
    }

    public static function getAvatarsByIds(array $userIds): array
    {
        $ids = array_values(array_unique(array_map('intval', $userIds)));
        $ids = array_values(array_filter($ids, fn ($v) => $v > 0));
        if (! $ids) {
            return [];
        }

        $keys = array_map(fn ($id) => self::cacheKey($id), $ids);
        $cached = Cache::many($keys);

        $out = [];
        $missIds = [];

        foreach ($ids as $i => $id) {
            $v = $cached[$keys[$i]] ?? null;
            if (is_array($v)) {
                $out[$id] = $v;
            } else {
                $missIds[] = $id;
            }
        }

        if ($missIds) {
            $rows = DB::table('zetawiki.user as A')
                ->leftJoin('profiles as B', 'A.user_id', '=', 'B.user_id')
                ->whereIn('A.user_id', $missIds)
                ->get(['A.user_id', 'A.user_name', 'B.t', 'B.ghash']);

            $toCache = [];

            foreach ($rows as $row) {
                $avatar = self::rowToAvatar($row);
                $id = $avatar['id'];

                $out[$id] = $avatar;
                $toCache[self::cacheKey($id)] = $avatar;
            }

            if ($toCache) {
                Cache::putMany($toCache, self::CACHE_TTL);
            }
        }

        return $out;
    }

    public static function forget(int $userId): void
    {
        if ($userId < 1) {
            return;
        }

        Cache::forget(self::cacheKey($userId));
    }
}
