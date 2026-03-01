<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class MwStatistics extends Model
{
    public $timestamps = false;
    protected $table = 'mw_statistics';
    protected $fillable = [
        'timeslot',
        'pages',
        'articles',
        'edits',
        'images',
        'users',
        'activeusers',
        'admins',
        'jobs',
    ];
    protected $casts = [
        'timeslot' => 'date',
        'pages' => 'integer',
        'articles' => 'integer',
        'edits' => 'integer',
        'images' => 'integer',
        'users' => 'integer',
        'activeusers' => 'integer',
        'admins' => 'integer',
        'jobs' => 'integer',
    ];
}
