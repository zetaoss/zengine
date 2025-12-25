<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Avatar extends Model
{
    protected $table = 'profiles';

    protected $primaryKey = 'user_id';

    public $incrementing = false;

    protected $keyType = 'int';

    public $timestamps = false;

    protected $fillable = [
        't',
        'ghash',
        'gravatar',
    ];

    protected $casts = [
        'user_id' => 'int',
        't' => 'int',
        'ghash' => 'string',
        'gravatar' => 'string',
    ];
}
