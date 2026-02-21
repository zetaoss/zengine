<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class CommonReport extends Model
{
    protected $appends = ['total'];
    protected $fillable = ['user_id', 'user_name', 'phase'];

    public function getTotalAttribute()
    {
        return $this->items->pluck('total')->sum();
    }

    public function items()
    {
        return $this->hasMany(CommonReportItem::class, 'report_id')
            ->orderBy('total', 'desc');
    }
}
