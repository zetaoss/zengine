<?php
namespace App\Jobs;

use App\Models\CommonReport;
use App\Services\SearchService;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Log;

class CommonReportJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    protected $report;
    public $tries = 1;

    public function __construct($report_id)
    {
        Log::debug("report_id: $report_id");
        $this->report = CommonReport::find($report_id);
    }

    public function handle(): void
    {
        $this->report->state = 1;
        $this->report->save();
        $this->processReport();
        $this->report->state = 2;
        $this->report->save();
    }

    private function processReport(): void
    {
        $searchService = new SearchService();
        $items = $this->report->items;

        foreach ($items as $item) {
            $word = $item->name;

            try {
                $item->daum_blog = $searchService->search('daum_blog', $word);
                $item->naver_blog = $searchService->search('naver_blog', $word);
                $item->naver_book = $searchService->search('naver_book', $word);
                $item->naver_news = $searchService->search('naver_news', $word);
                $item->google_search = $searchService->search('google_search', $word);

                $item->updateTotal();
                $item->save();
            } catch (\Exception $e) {
                Log::error("Failed to process item", [
                    'item_id' => $item->id,
                    'word' => $word,
                    'error' => $e->getMessage(),
                ]);
            }
        }

        Log::debug('processReport: done');
    }
}
