<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class StatGaHourly extends Model
{
    public const COLUMN_NAMES = [
        'active_users',
        'screen_page_views',
        'sessions',
    ];

    public $timestamps = false;
    protected $table = 'stat_ga_hourly';
    protected $fillable = [
        'timeslot',
        'active_users',
        'screen_page_views',
        'sessions',
    ];
    protected $casts = [
        'timeslot' => 'datetime',
        'active_users' => 'integer',
        'screen_page_views' => 'integer',
        'sessions' => 'integer',
    ];
}
