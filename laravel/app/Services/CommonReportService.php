<?php

namespace App\Services;

use App\Jobs\CommonReportJob;
use App\Models\CommonReport;
use App\Models\CommonReportItem;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Http;

class CommonReportService
{
    public function create(array $names, int $userId, string $userName): CommonReport
    {
        return DB::transaction(function () use ($names, $userId, $userName) {
            $report = CommonReport::create([
                'user_id' => $userId,
                'user_name' => $userName,
                'phase' => 'pending',
            ]);

            foreach ($names as $name) {
                CommonReportItem::create([
                    'report_id' => $report->id,
                    'name' => $name,
                ]);
            }

            $this->dispatchReportJobs($report->id);

            return $report;
        });
    }

    public function rerun(CommonReport $report): void
    {
        $report->update(['phase' => 'pending']);
        $this->dispatchReportJobs($report->id);
    }

    public function clone(CommonReport $original, int $userId, string $userName): CommonReport
    {
        return DB::transaction(function () use ($original, $userId, $userName) {
            $clone = CommonReport::create([
                'user_id' => $userId,
                'user_name' => $userName,
                'phase' => 'pending',
            ]);

            foreach ($original->items as $item) {
                CommonReportItem::create([
                    'report_id' => $clone->id,
                    'name' => $item->name,
                ]);
            }

            $this->dispatchReportJobs($clone->id);

            return $clone;
        });
    }

    public function delete(CommonReport $report): void
    {
        $report->delete();
    }

    public function processReport(CommonReport $report): void
    {
        $report->update(['phase' => 'running']);

        $response = Http::get($this->buildUrl($report));

        if (! $response->successful()) {
            throw new \RuntimeException('API request failed with status '.$response->status());
        }

        $this->updateReportData($report, $response->json());

        $report->update(['phase' => 'succeeded']);
    }

    private function dispatchReportJobs(int $reportId): void
    {
        CommonReportJob::dispatch($reportId);
    }

    private function buildUrl(CommonReport $report): string
    {
        $query = $report->items
            ->pluck('name')
            ->map(fn ($name) => 'q='.urlencode($name))
            ->implode('&');

        return rtrim(getenv('SEARCH_URL'), '/')."/search?$query";
    }

    private function updateReportData(CommonReport $report, array $data): void
    {
        $engines = $data['result']['engines'] ?? [];
        $values = $data['result']['values'] ?? [];

        $report->items->each(function ($item, $index) use ($engines, $values) {
            $dataMap = array_combine($engines, $values[$index] ?? []);

            $item->fill([
                'daum_blog' => $dataMap['daum_blog'] ?? 0,
                'naver_blog' => $dataMap['naver_blog'] ?? 0,
                'naver_book' => $dataMap['naver_book'] ?? 0,
                'naver_news' => $dataMap['naver_news'] ?? 0,
                'google_search' => $dataMap['google_search'] ?? 0,
            ]);

            $item->updateTotal();
            $item->save();
        });
    }
}
