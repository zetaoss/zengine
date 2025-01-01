<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class CommonReportItem extends Model
{
    protected $fillable = ['report_id', 'name', 'total', 'daum_blog', 'naver_blog', 'naver_book', 'naver_news', 'google_search'];

    public function report()
    {
        return $this->belongsTo('App\Models\CommonReport');
    }

    public function updateTotal()
    {
        $this->total = $this->daum_blog + $this->naver_blog + $this->naver_book + $this->naver_news + $this->google_search;
    }
}
