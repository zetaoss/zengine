<?php

namespace App\Http\Controllers;

use App\Models\User;
use Illuminate\Support\Facades\DB;

class UserController extends Controller
{
    public function show(string $userName)
    {
        $normalizedName = str_replace('_', ' ', $userName);

        $user = User::with('avatarRelation')
            ->where('user_name', $normalizedName)
            ->firstOrFail([
                'user_id',
                'user_name',
                'user_registration',
                'user_editcount',
            ]);

        return response()->json($user);
    }

    public function stats(int $userId)
    {
        $rows = DB::connection('mwdb')
            ->table('revision as r')
            ->join('actor as a', 'r.rev_actor', '=', 'a.actor_id')
            ->selectRaw('COUNT(*) as rev, DATE(r.rev_timestamp) as dt')
            ->where('a.actor_user', $userId)
            ->groupBy('dt')
            ->orderBy('dt')
            ->get();
        $data = [];
        foreach ($rows as $row) {
            $data[$row->dt] = (int) $row->rev;
        }

        return response()->json($data);
    }
}
