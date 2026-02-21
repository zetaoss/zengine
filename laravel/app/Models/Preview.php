<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Preview extends Model
{
    protected $fillable = ['code', 'type', 'title', 'image', 'description'];
    protected $hidden = ['url_md5', 'url', 'created_at', 'updated_at'];
    protected $primaryKey = 'url_md5';
    public $incrementing = false;
}
