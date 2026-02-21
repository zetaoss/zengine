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
use Throwable;

class CommonReportJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    public int $tries = 2;
    public int $timeout = 30;
    public bool $failOnTimeout = true;
    public int $backoff = 5;

    public function __construct(public int $reportId) {}

    public function handle(CommonReportService $service): void
    {
        $report = CommonReport::with('items')->find($this->reportId);

        if (! $report) {
            Log::warning("Report #{$this->reportId} not found");

            return;
        }

        $service->processReport($report);
        Log::info("Report #{$report->id} processed successfully.");
    }

    public function failed(Throwable $e): void
    {
        $report = CommonReport::find($this->reportId);
        if ($report && in_array($report->phase, ['pending', 'running'], true)) {
            $report->update(['phase' => 'failed']);
        }

        Log::error("Failed to process report #{$this->reportId}", [
            'error' => $e->getMessage(),
        ]);
    }
}
