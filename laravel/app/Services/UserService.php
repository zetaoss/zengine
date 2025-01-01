<?php
namespace App\Services;

use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;

class UserService
{
    public static function getUserAvatar(int $userID)
    {
        if ($userID == 0) {
            return ['id' => 0, 'name' => '?', 't' => 2, 'ghash' => ''];
        }
        $key = "userAvatar:$userID";
        $cached = Cache::get($key);
        if ($cached) {
            return $cached;
        }
        $mwdb = DB::connection('mwdb')->getDatabaseName();
        $rows = DB::select("SELECT A.user_name, B.t, B.ghash FROM $mwdb.user A LEFT JOIN profiles B ON A.user_id=B.user_id WHERE A.user_id=?", [$userID]);
        if (count($rows) == 0) {
            return null;
        }
        $user = $rows[0];
        $userAvatar = ['id' => $userID, 'name' => $user->user_name, 't' => $user->t, 'ghash' => $user->ghash];
        Cache::put($key, $userAvatar);
        return $userAvatar;
    }
}
