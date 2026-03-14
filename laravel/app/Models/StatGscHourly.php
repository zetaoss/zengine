<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class StatGscHourly extends Model
{
    public const COLUMN_NAMES = [
        'clicks',
        'impressions',
        'ctr',
        'position',
    ];

    public $timestamps = false;
    protected $table = 'stat_gsc_hourly';
    protected $fillable = [
        'timeslot',
        'clicks',
        'impressions',
        'ctr',
        'position',
    ];
    protected $casts = [
        'timeslot' => 'datetime',
        'clicks' => 'integer',
        'impressions' => 'integer',
        'ctr' => 'float',
        'position' => 'float',
    ];
}
