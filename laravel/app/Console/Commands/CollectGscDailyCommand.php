<?php

namespace App\Console\Commands;

use App\Services\Stat\CollectGscDailyService;
use Illuminate\Console\Command;
use RuntimeException;

class CollectGscDailyCommand extends Command
{
    protected $signature = 'z:collect-gsc-daily
                            {--days=10 : Number of trailing days to refresh (Search Console timezone)}
                            {--debug : Print raw Google Search Console JSON responses}';
    protected $description = 'Collect Google Search Console daily data';

    public function handle(CollectGscDailyService $daily): int
    {
        $days = (int) $this->option('days');
        $debug = (bool) $this->option('debug');
        $debugWriter = function (string $label, string $json): void {
            $this->line('');
            $this->info($label);
            $this->line($json);
        };

        try {
            $this->info('Running daily Google Search Console collection...');
            $result = $daily->collect($days, $debug, $debugWriter);
            $this->info("Site: {$result['siteUrl']}");
            $this->line("Timezone: {$result['timezone']}");
            $this->line('Since: '.($result['sinceLocal']->format('Y-m-d')));
            $this->line('Until: '.($result['untilLocal']->format('Y-m-d')));
            $this->line('Daily timeslots fetched: '.count((array) ($result['timeslots'] ?? [])));
            $this->line(
                "Daily DB write: inserted={$result['db']['inserted']}, updated={$result['db']['updated']}, skipped={$result['db']['skipped']}"
            );
            $this->line('');
            $this->info('daily Google Search Console collection completed.');

            return Command::SUCCESS;
        } catch (RuntimeException $e) {
            $this->error($e->getMessage());

            return Command::FAILURE;
        }
    }
}
