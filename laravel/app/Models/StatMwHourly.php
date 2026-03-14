<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class StatMwHourly extends Model
{
    public $timestamps = false;
    protected $table = 'stat_mw_hourly';
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
        'timeslot' => 'datetime',
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
