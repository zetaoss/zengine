<?php

namespace App\Console\Commands;

use App\Services\Stat\CollectGscHourlyService;
use Illuminate\Console\Command;
use RuntimeException;

class CollectGscHourlyCommand extends Command
{
    protected $signature = 'z:collect-gsc-hourly
                            {--days=2 : Number of trailing days from current Search Console hour}
                            {--debug : Print raw Google Search Console JSON responses}';
    protected $description = 'Collect Google Search Console hourly data';

    public function handle(CollectGscHourlyService $hourly): int
    {
        $days = (int) $this->option('days');
        $debug = (bool) $this->option('debug');
        $debugWriter = function (string $label, string $json): void {
            $this->line('');
            $this->info($label);
            $this->line($json);
        };

        try {
            $this->info('Running hourly Google Search Console collection...');
            $result = $hourly->collect($days, $debug, $debugWriter);
            $this->info("Site: {$result['siteUrl']}");
            $this->line("Timezone: {$result['timezone']}");
            $this->line('Since: '.($result['sinceLocal']->format('Y-m-d\TH:i:s')));
            $this->line('Until: '.($result['untilLocal']->format('Y-m-d\TH:i:s')));
            $this->line('Hourly timeslots fetched: '.count((array) ($result['timeslots'] ?? [])));
            $this->line(
                "Hourly DB write: inserted={$result['db']['inserted']}, updated={$result['db']['updated']}, skipped={$result['db']['skipped']}"
            );
            $this->line('');
            $this->info('hourly Google Search Console collection completed.');

            return Command::SUCCESS;
        } catch (RuntimeException $e) {
            $this->error($e->getMessage());

            return Command::FAILURE;
        }
    }
}
