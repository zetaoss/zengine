<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class User extends Model
{
    protected $table = 'zetawiki.user';

    protected $primaryKey = 'user_id';

    public $timestamps = false;

    protected $appends = ['avatar'];

    protected $hidden = ['avatarRelation'];

    public function avatarRelation()
    {
        return $this->hasOne(Avatar::class, 'user_id');
    }

    public function getAvatarAttribute(): array
    {
        $avatar = $this->avatarRelation;

        return [
            'id' => $this->user_id,
            'name' => $this->user_name,
            't' => (int) ($avatar?->t ?? 1),
            'gravatar' => (string) ($avatar?->gravatar ?? ''),
            'ghash' => (string) ($avatar?->ghash ?? ''),
        ];
    }
}
