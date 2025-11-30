<?php

namespace App\Http\Controllers;

use App\Services\AvatarService;
use Illuminate\Support\Facades\DB;

class AvatarController extends Controller
{
    public function show(string $username)
    {
        $username = trim($username);

        if ($username === '' || $username === '?') {
            return response()->json([
                'id' => 0,
                'name' => '?',
                't' => 2,
                'ghash' => '',
            ]);
        }

        $mwdb = DB::connection('mwdb')->getDatabaseName();
        $row = DB::selectOne("SELECT user_id FROM $mwdb.user WHERE user_name = ?", [$username]);

        if (! $row) {
            return response()->json([
                'error' => 'User not found',
            ], 404);
        }

        $userID = (int) $row->user_id;

        $avatar = AvatarService::getAvatarById($userID);

        if ($avatar === null) {
            return response()->json([
                'error' => 'Avatar not found',
            ], 404);
        }

        return response()->json($avatar);
    }
}
