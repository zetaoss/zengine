<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class PageComment extends Model
{
    protected $table = 'zetawiki.page_comments';

    public $timestamps = false;

    protected $fillable = [
        'page_id',
        'message',
        'user_id',
        'user_name',
        'created',
    ];

    protected $casts = [
        'id' => 'int',
        'page_id' => 'int',
        'user_id' => 'int',
        'created' => 'datetime',
    ];

    public function page()
    {
        return $this->belongsTo(Page::class, 'page_id', 'page_id');
    }
}
