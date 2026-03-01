<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class CfAnalyticsHourly extends Model
{
    public $timestamps = false;
    protected $table = 'cf_analytics_hourly';

    protected $fillable = [
        'timeslot',
        'name',
        'value',
    ];

    protected $casts = [
        'timeslot' => 'datetime',
    ];
}
