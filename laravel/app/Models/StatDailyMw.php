<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class StatDailyMw extends Model
{
    public $timestamps = false;
    protected $table = 'stat_daily_mw';
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
