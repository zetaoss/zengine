<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class WriteRequest extends Model
{
    protected $fillable = [
        'title',
        'user_id',
        'user_name',
        'rate',
        'ref',
        'hit',
    ];
}
