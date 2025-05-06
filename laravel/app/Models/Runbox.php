<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Runbox extends Model
{
    protected $fillable = ['hash', 'phase', 'user_id', 'page_id', 'type', 'payload', 'outs', 'cpu', 'mem', 'time'];

    protected $hidden = ['id', 'payload', 'created_at', 'updated_at'];

    protected $casts = ['payload' => 'array', 'outs' => 'array'];
}
