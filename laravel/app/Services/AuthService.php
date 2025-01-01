<?php
namespace App\Services;

use App\Services\UserService;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Redis;

class AuthService
{
    public static function me()
    {
        $userID = self::getValidUserID();
        if ($userID === false) {
            return false;
        }
        return self::getMeData($userID);
    }

    private static function getValidUserID(): int
    {
        if (!array_key_exists('zetawikiUserID', $_COOKIE) || !array_key_exists('zetawikiUserName', $_COOKIE)) {
            return false;
        }
        $userID = $_COOKIE['zetawikiUserID'];
        $userName = $_COOKIE['zetawikiUserName'];

        if (array_key_exists('zetawiki_session', $_COOKIE)) {
            $redis = Redis::connection('mwsession');
            $value = $redis->get('zetawiki:MWSession:' . $_COOKIE['zetawiki_session']);
            if ($value) {
                $arr = unserialize($value);
                $wsUserID = $arr['data']['wsUserID'] ?? 0;
                if ($wsUserID == $userID) {
                    return $userID;
                }
            }
        }

        if (array_key_exists('zetawikiToken', $_COOKIE)) {
            $cnt = count(DB::select('SELECT user_id FROM zetawiki.user WHERE user_id=? AND user_name=? AND user_token=?', [$userID, $userName, $_COOKIE['zetawikiToken']]));
            if ($cnt == 1) {
                return $userID;
            }
        }
        return false;
    }

    private static function getMeData(int $userID)
    {
        $key = "meData:$userID";
        Cache::delete($key);
        $cached = Cache::get($key);
        if ($cached) {
            return $cached;
        }
        $mwdb = DB::connection('mwdb')->getDatabaseName();
        $rows = DB::select("SELECT group_concat(ug_group) groups FROM $mwdb.user_groups WHERE ug_user=?", [$userID]);
        $groups = [];
        if (count($rows) == 1 && !empty($rows[0]->groups)) {
            $groups = explode(',', $rows[0]->groups);
        }
        $meData = ['avatar' => UserService::getUserAvatar($userID), 'groups' => $groups];
        Cache::put($key, $meData);
        return $meData;
    }
}
