<?php

namespace App\Services;

use App\Jobs\CommonReportJob;
use App\Models\CommonReport;
use App\Models\CommonReportItem;
use Illuminate\Support\Facades\DB;

class CommonReportService
{
    public function create(array $names, int $userId): CommonReport
    {
        return DB::transaction(function () use ($names, $userId) {
            $report = CommonReport::create([
                'user_id' => $userId,
                'phase' => 'pending',
            ]);

            foreach ($names as $name) {
                CommonReportItem::create([
                    'report_id' => $report->id,
                    'name' => $name,
                ]);
            }

            CommonReportJob::dispatch($report->id);

            return $report;
        });
    }

    public function rerun(CommonReport $report): void
    {
        $report->update(['phase' => 'pending']);
        CommonReportJob::dispatch($report->id);
    }

    public function clone(CommonReport $original, int $userId): CommonReport
    {
        return DB::transaction(function () use ($original, $userId) {
            $clone = CommonReport::create([
                'user_id' => $userId,
                'phase' => 'pending',
            ]);

            foreach ($original->items as $item) {
                CommonReportItem::create([
                    'report_id' => $clone->id,
                    'name' => $item->name,
                ]);
            }

            CommonReportJob::dispatch($clone->id);

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

        $query = $report->items
            ->pluck('name')
            ->map(fn ($name) => 'q='.urlencode($name))
            ->implode('&');

        $url = rtrim(getenv('SEARCH_URL'), '/')."/search?$query";

        $response = Http::get($url);

        if (! $response->successful()) {
            throw new RuntimeException('API request failed with status '.$response->status());
        }

        $data = $response->json();
        $engines = $data['result']['engines'] ?? [];
        $values = $data['result']['values'] ?? [];

        $report->items->each(function ($item, $index) use ($engines, $values) {
            $row = $values[$index] ?? [];
            $map = array_combine($engines, $row) ?: [];

            $item->fill([
                'daum_blog' => $map['daum_blog'] ?? 0,
                'naver_blog' => $map['naver_blog'] ?? 0,
                'naver_book' => $map['naver_book'] ?? 0,
                'naver_news' => $map['naver_news'] ?? 0,
                'google_search' => $map['google_search'] ?? 0,
            ]);

            $item->updateTotal();
            $item->save();
        });

        $report->update(['phase' => 'succeeded']);
    }

    public function markTimedOutReports(\DateTimeImmutable $threshold): int
    {
        $updatedCount = 0;

        CommonReport::whereIn('phase', ['pending', 'running'])
            ->where('created_at', '<', $threshold)
            ->chunk(100, function ($reports) use (&$updatedCount) {
                foreach ($reports as $report) {
                    $report->update(['phase' => 'failed']);
                    $updatedCount++;
                }
            });

        return $updatedCount;
    }
}
