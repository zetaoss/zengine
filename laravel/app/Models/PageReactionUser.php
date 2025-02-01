<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class PageReactionUser extends Model
{
    public $timestamps = false;

    protected $fillable = ['page_id', 'user_id', 'emoji_code'];
}
