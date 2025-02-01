<?php

namespace App\Models;

use App\Services\UserService;
use DB;
use Illuminate\Database\Eloquent\Model;

class WriteRequest extends Model
{
    protected $attributes = ['rate' => 0];

    protected $appends = ['hit', 'userAvatar'];

    public function getUserAvatarAttribute()
    {
        return UserService::getUserAvatar($this->user_id);
    }

    public function getHitAttribute()
    {
        $row = DB::table('not_matches')->where('title', '=', $this->title)->first();
        if ($row) {
            return $row->hit;
        }

        return 0;
    }
}
