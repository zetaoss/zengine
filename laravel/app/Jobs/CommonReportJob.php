<?php

namespace App\Jobs;

use App\Models\CommonReport;
use App\Services\CommonReportService;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Log;

class CommonReportJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    public int $tries = 3;

    public function __construct(public int $reportId) {}

    public function handle(CommonReportService $service): void
    {
        $report = CommonReport::with('items')->find($this->reportId);

        if (! $report) {
            Log::warning("Report #{$this->reportId} not found");

            return;
        }

        try {
            $service->processReport($report);
            Log::info("Report #{$report->id} processed successfully.");
        } catch (\Throwable $e) {
            Log::error("Failed to process report #{$report->id}", [
                'error' => $e->getMessage(),
            ]);
            $report->update(['phase' => 'failed']);
        }
    }
}
