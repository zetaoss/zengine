<?php

namespace App\Console\Commands;

use App\Services\Stat\CollectGaDailyService;
use Illuminate\Console\Command;
use RuntimeException;

class CollectGaDailyCommand extends Command
{
    protected $signature = 'z:collect-ga-daily
                            {--days=7 : Number of trailing days to refresh (today included, property timezone)}
                            {--debug : Print raw Google Analytics JSON responses}';
    protected $description = 'Collect Google Analytics daily data';

    public function handle(CollectGaDailyService $daily): int
    {
        $days = (int) $this->option('days');
        $debug = (bool) $this->option('debug');
        $debugWriter = function (string $label, string $json): void {
            $this->line('');
            $this->info($label);
            $this->line($json);
        };

        try {
            $this->info('Running daily Google Analytics collection...');
            $result = $daily->collect($days, $debug, $debugWriter);
            $this->info("Property: {$result['propertyId']}");
            $this->line("Timezone: {$result['timezone']}");
            $this->line('Since: '.($result['sinceLocal']->format('Y-m-d')));
            $this->line('Until: '.($result['untilLocal']->format('Y-m-d')));
            $this->line('Daily timeslots fetched: '.count((array) ($result['timeslots'] ?? [])));
            $this->line(
                "Daily DB write: inserted={$result['db']['inserted']}, updated={$result['db']['updated']}, skipped={$result['db']['skipped']}"
            );
            $this->line('');
            $this->info('daily Google Analytics collection completed.');

            return Command::SUCCESS;
        } catch (RuntimeException $e) {
            $this->error($e->getMessage());

            return Command::FAILURE;
        }
    }
}
