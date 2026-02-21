<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class UserSocial extends Model
{
    protected $table = 'zetawiki.user_social';

    protected $primaryKey = 'id';

    public $timestamps = false;

    protected $fillable = [
        'provider',
        'social_id',
        'user_id',
        'deletion_code',
    ];

    protected $casts = [
        'id' => 'int',
        'user_id' => 'int',
    ];
}
