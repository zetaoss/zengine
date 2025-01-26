<?php
namespace App\Jobs;

use App\Models\CommonReport;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;

class CommonReportJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    protected CommonReport $report;
    public int $tries = 1;

    public function __construct(int $id)
    {
        Log::debug("Initializing job with report ID: $id");
        $this->report = CommonReport::findOrFail($id);
    }

    public function handle(): void
    {
        $this->updateState(1);

        try {
            $response = Http::get($this->buildUrl());

            if ($response->successful()) {
                $this->updateReportData($response->json());
            } else {
                Log::error('API request failed', [
                    'status' => $response->status(),
                    'body'   => $response->body(),
                ]);
            }
        } catch (\Exception $e) {
            Log::error('API request error', ['error' => $e->getMessage()]);
        } finally {
            $this->updateState(2);
        }

        Log::debug('Report processing complete');
    }

    private function updateState(int $state): void
    {
        $this->report->update(['state' => $state]);
    }

    private function buildUrl(): string
    {
        $query = $this->report->items
            ->pluck('name')
            ->map(fn($name) => 'q=' . urlencode($name))
            ->implode('&');

        return getenv("SEARCH_URL") . "/search?$query";
    }

    private function updateReportData(array $data): void
    {
        $engines = $data['result']['engines'] ?? [];
        $values  = $data['result']['values'] ?? [];

        $this->report->items->each(function ($item, $index) use ($engines, $values) {
            $dataMap = array_combine($engines, $values[$index] ?? []);

            $item->fill([
                'daum_blog'     => $dataMap['daum_blog'] ?? 0,
                'naver_blog'    => $dataMap['naver_blog'] ?? 0,
                'naver_book'    => $dataMap['naver_book'] ?? 0,
                'naver_news'    => $dataMap['naver_news'] ?? 0,
                'google_search' => $dataMap['google_search'] ?? 0,
            ]);

            $item->updateTotal();
            $item->save();
        });
    }
}
