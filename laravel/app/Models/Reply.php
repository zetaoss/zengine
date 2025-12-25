<?php

namespace App\Models;

use App\Services\AvatarService;
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
        return AvatarService::getAvatarById((int) $this->user_id);
    }
}
