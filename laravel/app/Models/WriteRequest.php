<?php

namespace App\Models;

use App\Services\AvatarService;
use DB;
use Illuminate\Database\Eloquent\Model;

class WriteRequest extends Model
{
    protected $attributes = ['rate' => 0];

    protected $appends = ['hit', 'avatar'];

    public function getAvatarAttribute(): ?array
    {
        return AvatarService::getAvatarById((int) $this->user_id);
    }

    public function getHitAttribute(): int
    {
        $row = DB::table('not_matches')->where('title', '=', $this->title)->first();

        return $row ? (int) $row->hit : 0;
    }
}
