<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Tag extends Model
{
    protected $fillable = ['name', 'wiki_title', 'wiki_cat'];
    public function posts()
    {
        return $this->belongsToMany(Post::class);
    }
}
