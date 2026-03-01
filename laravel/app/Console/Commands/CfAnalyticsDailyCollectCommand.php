<?php

namespace App\Console\Commands;

use App\Services\CfAnalytics\CfAnalyticsDailyCollectorService;
use Illuminate\Console\Command;
use RuntimeException;

class CfAnalyticsDailyCollectCommand extends Command
{
    protected $signature = 'app:cf-analytics-daily-collect
                            {--days=7 : Number of days to include (today included, UTC)}
                            {--debug : Print raw GraphQL JSON responses}';

    protected $description = 'Collect Cloudflare analytics daily dataset';

    public function handle(CfAnalyticsDailyCollectorService $daily): int
    {
        $days = (int) $this->option('days');
        $debug = (bool) $this->option('debug');
        $debugWriter = function (string $label, string $json): void {
            $this->line('');
            $this->info($label);
            $this->line($json);
        };

        try {
            $this->info('Running daily analytics collection...');
            $dailyResult = $daily->collect($days, $debug, $debugWriter);
            $this->info("Zone: {$dailyResult['zoneId']}");
            $this->line('Since: '.($dailyResult['sinceUtc']->utc()->format('Y-m-d')));
            $this->line('Until: '.($dailyResult['untilUtc']->utc()->format('Y-m-d')));
            $this->line('Daily timeslots fetched: '.count((array) ($dailyResult['timeslots'] ?? [])));
            $this->line(
                "Daily DB write: inserted={$dailyResult['db']['inserted']}, updated={$dailyResult['db']['updated']}, skipped={$dailyResult['db']['skipped']}"
            );
            $this->line('');
            $this->info('daily analytics collection completed.');

            return Command::SUCCESS;
        } catch (RuntimeException $e) {
            $this->error($e->getMessage());

            return Command::FAILURE;
        }
    }
}
