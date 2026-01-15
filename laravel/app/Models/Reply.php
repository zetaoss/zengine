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

    public function post()
    {
        return $this->belongsTo(Post::class, 'post_id');
    }
}
