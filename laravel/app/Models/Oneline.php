<?php

namespace App\Models;

use App\Services\UserService;
use Illuminate\Database\Eloquent\Model;

class Oneline extends Model
{
    protected $appends = ['userAvatar'];

    public function getUserAvatarAttribute()
    {
        return UserService::getUserAvatar($this->user_id);
    }
}
