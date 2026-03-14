<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class StatHourlyCf extends Model
{
    public $timestamps = false;
    protected $table = 'stat_hourly_cf';
    protected $fillable = [
        'timeslot',
        'name',
        'value',
    ];
    protected $casts = [
        'timeslot' => 'datetime',
    ];
}
