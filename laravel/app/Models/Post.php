<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Post extends Model
{
    protected $fillable = [
        'user_id',
        'user_name',
        'cat',
        'title',
        'body',
        'is_notice',
        'tags_str',
        'hit',
        'channel_id',
    ];

    protected $casts = [
        'user_id' => 'integer',
        'channel_id' => 'integer',
        'hit' => 'integer',
        'is_notice' => 'boolean',
    ];

    protected $appends = ['tag_names'];

    public function replies()
    {
        return $this->hasMany(Reply::class, 'post_id');
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
