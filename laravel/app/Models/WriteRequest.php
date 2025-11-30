<?php

namespace App\Models;

use App\Services\AvatarService;
use DB;
use Illuminate\Database\Eloquent\Model;

class WriteRequest extends Model
{
    protected $attributes = ['rate' => 0];

    protected $appends = ['hit', 'avatar'];

    public function getAvatarAttribute()
    {
        return AvatarService::getAvatarById($this->user_id);
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
