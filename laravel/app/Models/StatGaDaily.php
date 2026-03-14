<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class StatGaDaily extends Model
{
    public const COLUMN_NAMES = [
        'active_users',
        'screen_page_views',
        'sessions',
    ];

    public $timestamps = false;
    protected $table = 'stat_ga_daily';
    protected $fillable = [
        'timeslot',
        'active_users',
        'screen_page_views',
        'sessions',
    ];
    protected $casts = [
        'timeslot' => 'date',
        'active_users' => 'integer',
        'screen_page_views' => 'integer',
        'sessions' => 'integer',
    ];
}
