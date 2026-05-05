<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class DocTask extends Model
{
    protected $fillable = [
        'user_id',
        'user_name',
        'title',
        'request_type',
        'content',
        'llm_model',
        'phase',
        'attempts',
        'error_count',
        'skip_count',
        'last_error',
    ];
}
