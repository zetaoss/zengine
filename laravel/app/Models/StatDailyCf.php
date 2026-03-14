<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class StatDailyCf extends Model
{
    public $timestamps = false;
    protected $table = 'stat_daily_cf';
    protected $fillable = [
        'timeslot',
        'name',
        'value',
    ];
    protected $casts = [
        'timeslot' => 'date',
    ];
}
