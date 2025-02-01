<?php

namespace App\Models;

use App\Services\AuthService;
use App\Services\UserService;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

class Reply extends Model
{
    use SoftDeletes;

    protected $dates = ['deleted_at'];

    protected $fillable = ['user_id', 'body'];

    protected $appends = ['userAvatar'];

    public function getUserAvatarAttribute()
    {
        return UserService::getUserAvatar($this->user_id);
    }

    protected static function boot()
    {
        parent::boot();
        static::created(function ($reply) {
            $reply->post->increment('replies_count');
        });
        static::deleted(function ($reply) {
            $me = AuthService::me();
            if (! $me) {
                return;
            }
            $reply->delete_user_id = $me['avatar']['id'];
            $reply->save();
            $reply->post->update(['replies_count' => $reply->post->replies()->count()]);
        });
    }

    public function post()
    {
        return $this->belongsTo(Post::class, 'post_id');
    }
}
