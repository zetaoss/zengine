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
        'user_id',
        'social_id',
        'deauthorized_at',
        'deleted_at',
        'deletion_code',
    ];
    protected $casts = [
        'id' => 'int',
        'user_id' => 'int',
        'deauthorized_at' => 'datetime',
        'deleted_at' => 'datetime',
    ];
}
