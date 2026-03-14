<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class StatGscDaily extends Model
{
    public const COLUMN_NAMES = [
        'clicks',
        'impressions',
        'ctr',
        'position',
    ];

    public $timestamps = false;
    protected $table = 'stat_gsc_daily';
    protected $fillable = [
        'timeslot',
        'clicks',
        'impressions',
        'ctr',
        'position',
    ];
    protected $casts = [
        'timeslot' => 'date',
        'clicks' => 'integer',
        'impressions' => 'integer',
        'ctr' => 'float',
        'position' => 'float',
    ];
}
