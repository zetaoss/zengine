<?php

namespace App\Models;

use DB;
use Illuminate\Database\Eloquent\Model;

class WriteRequest extends Model
{
    protected $attributes = ['rate' => 0];

    protected $appends = ['hit'];

    public function getHitAttribute(): int
    {
        $row = DB::table('not_matches')->where('title', '=', $this->title)->first();

        return $row ? (int) $row->hit : 0;
    }
}
