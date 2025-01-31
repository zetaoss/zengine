<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Runbox extends Model
{
    protected $hidden = ['id', 'payload', 'created_at', 'updated_at'];
    protected $casts  = ['payload' => 'array', 'outs' => 'array'];
}
