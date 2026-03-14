<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class StatCfHourly extends Model
{
    public $timestamps = false;
    protected $table = 'stat_cf_hourly';
    protected $fillable = [
        'timeslot',
        'name',
        'value',
    ];
    protected $casts = [
        'timeslot' => 'datetime',
    ];
}
