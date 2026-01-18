<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Oneline extends Model
{
    public $timestamps = false;

    protected $fillable = [
        'user_id',
        'user_name',
        'created',
        'message',
    ];

    protected $casts = [
        'id' => 'int',
        'user_id' => 'int',
    ];
}
