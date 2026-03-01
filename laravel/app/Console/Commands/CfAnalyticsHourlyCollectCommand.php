<?php

namespace App\Console\Commands;

use App\Services\CfAnalytics\CfAnalyticsHourlyCollectorService;
use Illuminate\Console\Command;
use RuntimeException;

class CfAnalyticsHourlyCollectCommand extends Command
{
    protected $signature = 'app:cf-analytics-hourly-collect
                            {--days=2 : Number of trailing days from current UTC hour (hour-aligned)}
                            {--debug : Print raw GraphQL JSON responses}';

    protected $description = 'Collect Cloudflare analytics hourly dataset';

    public function handle(CfAnalyticsHourlyCollectorService $hourly): int
    {
        $days = (int) $this->option('days');
        $debug = (bool) $this->option('debug');
        $debugWriter = function (string $label, string $json): void {
            $this->line('');
            $this->info($label);
            $this->line($json);
        };

        try {
            $this->info('Running hourly analytics collection...');
            $hourlyResult = $hourly->collect($days, $debug, $debugWriter);
            $this->info("Zone: {$hourlyResult['zoneId']}");
            $this->line('Since: '.($hourlyResult['sinceUtc']->utc()->format('Y-m-d\TH:i:s\Z')));
            $this->line('Until: '.($hourlyResult['untilUtc']->utc()->format('Y-m-d\TH:i:s\Z')));
            $this->line('Hourly timeslots fetched: '.count((array) ($hourlyResult['timeslots'] ?? [])));
            $this->line(
                "Hourly DB write: inserted={$hourlyResult['db']['inserted']}, updated={$hourlyResult['db']['updated']}, skipped={$hourlyResult['db']['skipped']}"
            );
            $this->line('');
            $this->info('hourly analytics collection completed.');

            return Command::SUCCESS;
        } catch (RuntimeException $e) {
            $this->error($e->getMessage());

            return Command::FAILURE;
        }
    }
}
