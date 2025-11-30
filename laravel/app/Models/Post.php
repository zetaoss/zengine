<?php

namespace App\Models;

use App\Services\AvatarService;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

class Post extends Model
{
    use SoftDeletes;

    protected $appends = ['avatar', 'tag_names'];

    protected $dates = ['deleted_at'];

    protected $hidden = ['channel_id', 'deleted_at', 'delete_user_id'];

    protected $fillable = ['user_id', 'cat', 'title', 'body', 'is_notice', 'tags_str', 'replies_count'];

    public function getAvatarAttribute()
    {
        return AvatarService::getAvatarById($this->user_id);
    }

    public function getTagNamesAttribute()
    {
        return explode(',', $this->tags_str);
    }

    protected static function boot()
    {
        parent::boot();
        static::deleting(function ($post) {
            $me = AvatarService::me();
            if (! $me) {
                return;
            }

            $post->delete_user_id = $me['id'];
            $post->save();
            $post->replies()->delete();
        });
    }

    public function save(array $options = [])
    {
        parent::save($options);
        $this->syncTags();
    }

    public function replies()
    {
        return $this->hasMany(Reply::class);
    }

    public function tags()
    {
        return $this->belongsToMany(Tag::class);
    }

    private function syncTags()
    {
        $tag_ids = [];
        foreach ($this->tag_names as $tag_name) {
            $tag = Tag::where('name', $tag_name)->first();
            if (is_null($tag)) {
                continue;
            }

            $tag_ids[] = $tag->id;
        }
        $this->tags()->sync($tag_ids);
    }

    public function createReply($reply)
    {
        return $this->replies()->create($reply);
    }
}
