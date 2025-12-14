<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Reply extends Model
{
    protected $fillable = [
        'post_id',
        'user_id',
        'body',
    ];

    protected $appends = ['avatar'];

    public function post()
    {
        return $this->belongsTo(Post::class, 'post_id');
    }

    public function getAvatarAttribute(): ?array
    {
        return \App\Services\AvatarService::getAvatarById((int) $this->user_id);
    }
}
