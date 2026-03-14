<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class StatCfDaily extends Model
{
    public $timestamps = false;
    protected $table = 'stat_cf_daily';
    protected $fillable = [
        'timeslot',
        'name',
        'value',
    ];
    protected $casts = [
        'timeslot' => 'date',
    ];
}
