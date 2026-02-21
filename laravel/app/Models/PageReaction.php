<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class PageReaction extends Model
{
    public $timestamps = false;
    protected $casts = ['emoji_count' => 'json'];
    protected $fillable = ['page_id', 'cnt', 'emoji_count'];
    protected $primaryKey = 'page_id';
    public $incrementing = false;
}
