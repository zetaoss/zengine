<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class CfAnalyticsDaily extends Model
{
    public $timestamps = false;
    protected $table = 'cf_analytics_daily';

    protected $fillable = [
        'timeslot',
        'name',
        'value',
    ];

    protected $casts = [
        'timeslot' => 'date',
    ];
}
