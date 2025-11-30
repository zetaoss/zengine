<?php

namespace App\Models;

use App\Services\AvatarService;
use Illuminate\Database\Eloquent\Model;

class CommonReport extends Model
{
    protected $appends = ['avatar', 'total'];

    protected $fillable = ['user_id', 'phase'];

    public function getAvatarAttribute()
    {
        return AvatarService::getAvatarById($this->user_id);
    }

    public function getTotalAttribute()
    {
        return $this->items->pluck('total')->sum();
    }

    protected static function boot()
    {
        parent::boot();
        static::deleting(function ($report) {
            $report->items->each->delete();
        });
    }

    public function items()
    {
        return $this->hasMany(CommonReportItem::class, 'report_id')->orderBy('total', 'desc');
    }

    public function addItem($item)
    {
        return $this->items()->create($item);
    }
}
