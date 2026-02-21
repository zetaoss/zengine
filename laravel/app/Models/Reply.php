<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Reply extends Model
{
    protected $fillable = [
        'post_id',
        'user_id',
        'user_name',
        'body',
    ];
    protected $casts = [
        'post_id' => 'integer',
        'user_id' => 'integer',
    ];

    public function post()
    {
        return $this->belongsTo(Post::class, 'post_id');
    }
}
