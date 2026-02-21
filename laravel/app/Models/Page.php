<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Page extends Model
{
    protected $table = 'zetawiki.page';
    protected $primaryKey = 'page_id';
    public $timestamps = false;
    protected $casts = [
        'page_id' => 'int',
    ];
}
