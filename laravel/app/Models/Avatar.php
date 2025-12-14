<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Avatar extends Model
{
    protected $table = 'profiles';

    protected $primaryKey = 'user_id';

    public $timestamps = false;

    public $incrementing = false;

    protected $keyType = 'int';

    protected $fillable = [
        't',
        'gravatar',
        'ghash',
    ];
}
