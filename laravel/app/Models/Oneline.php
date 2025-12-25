<?php

namespace App\Models;

use App\Services\AvatarService;
use Illuminate\Database\Eloquent\Model;

class Oneline extends Model
{
    protected $appends = ['avatar'];

    public function getAvatarAttribute(): ?array
    {
        return AvatarService::getAvatarById((int) $this->user_id);
    }
}
