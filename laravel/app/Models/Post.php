<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Post extends Model
{
    protected $fillable = [
        'user_id',
        'cat',
        'title',
        'body',
        'is_notice',
        'tags_str',
        'hit',
        'channel_id',
    ];

    protected $casts = [
        'is_notice' => 'boolean',
    ];

    protected $appends = ['avatar', 'tag_names'];

    public function replies()
    {
        return $this->hasMany(Reply::class, 'post_id');
    }

    public function getAvatarAttribute(): ?array
    {
        return \App\Services\AvatarService::getAvatarById((int) $this->user_id);
    }

    public function getTagNamesAttribute(): array
    {
        $str = (string) ($this->tags_str ?? '');
        if ($str === '') {
            return [];
        }

        return array_values(array_filter(array_map('trim', explode(',', $str))));
    }
}
