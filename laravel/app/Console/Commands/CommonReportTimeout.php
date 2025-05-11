<?php

namespace App\Console\Commands;

use App\Services\CommonReportService;
use DateTimeImmutable;
use Illuminate\Console\Command;

class CommonReportTimeout extends Command
{
    protected $signature = 'app:common-report-timeout';

    protected $description = 'Mark reports as failed if not processed within 1 minute';

    public function __construct(protected CommonReportService $service)
    {
        parent::__construct();
    }

    public function handle(): int
    {
        $this->info('['.$this->getName().'] Checking for timed-out reports...');

        $threshold = new DateTimeImmutable('-1 minute');
        $count = $this->service->markTimedOutReports($threshold);

        if ($count === 0) {
            $this->info('['.$this->getName().'] No timed-out reports found.');
        } else {
            $this->info('['.$this->getName()."] Total {$count} report(s) updated.");
        }

        return Command::SUCCESS;
    }
}
