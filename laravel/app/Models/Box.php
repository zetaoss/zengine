<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Box extends Model
{
    protected $hidden = ['id', 'curid', 'hash', 'api', 'lang', 'source', 'created_at', 'updated_at'];
    protected $casts = ['source' => 'array', 'metadata' => 'array', 'outs' => 'array'];
}
