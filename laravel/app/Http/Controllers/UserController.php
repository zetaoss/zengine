<?php

namespace App\Http\Controllers;

use App\Models\User;
use App\Services\AvatarService;
use Illuminate\Support\Facades\DB;

class UserController extends Controller
{
    public function show(string $userName)
    {
        $normalizedName = str_replace('_', ' ', $userName);

        $user = User::where('user_name', $normalizedName)
            ->firstOrFail([
                'user_id',
                'user_name',
                'user_registration',
                'user_editcount',
            ]);

        $payload = $user->toArray();
        $payload['avatar'] = AvatarService::getAvatarById((int) $user->user_id);

        return response()->json($payload);
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
