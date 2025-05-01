<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;

class CommonReportTimeout extends Command
{
    protected $signature = 'app:common-report-timeout';

    protected $description = 'Mark reports as failed if not processed within 1 minute';

    public function handle()
    {
        $threshold = Carbon::now()->subMinute();

        $timedOut = CommonReport::where('state', 0)
            ->where('created_at', '<', $threshold)
            ->get();

        if ($timedOut->isEmpty()) {
            $this->info('No timed-out reports found.');

            return 0;
        }

        foreach ($timedOut as $report) {
            $report->state = -1;
            $report->save();
            $this->info("Report #{$report->id} marked as failed.");
        }

        $this->info("Total {$timedOut->count()} report(s) updated.");

        return 0;
    }
}
