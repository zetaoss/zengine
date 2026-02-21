<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class User extends Model
{
    protected $table = 'zetawiki.user';
    protected $primaryKey = 'user_id';
    public $timestamps = false;
    protected $casts = [
        'user_id' => 'int',
    ];
}
