<?php

namespace App\Models;

use App\Services\AvatarService;
use Illuminate\Database\Eloquent\Model;

class CommonReport extends Model
{
    protected $appends = ['total', 'avatar'];

    protected $fillable = ['user_id', 'phase'];

    public function getTotalAttribute()
    {
        return $this->items->pluck('total')->sum();
    }

    public function getAvatarAttribute()
    {
        return AvatarService::getAvatarById((int) $this->user_id);
    }

    public function items()
    {
        return $this->hasMany(CommonReportItem::class, 'report_id')
            ->orderBy('total', 'desc');
    }
}
