<?php

namespace App\Console\Commands;

use App\Services\Stat\CollectMwDailyService;
use Illuminate\Console\Command;
use RuntimeException;

class CollectMwDailyCommand extends Command
{
    protected $signature = 'z:collect-mw-daily
                            {--date= : Target KST date (YYYY-MM-DD), default is today KST}
                            {--debug : Print raw JSON payload}';
    protected $description = 'Collect MediaWiki daily data';

    public function handle(CollectMwDailyService $daily): int
    {
        try {
            $result = $daily->collect(trim((string) $this->option('date')));

            if ((bool) $this->option('debug')) {
                $this->line(json_encode($result['payload'], JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES) ?: '{}');
            }

            $this->info("Stored stat_daily_mw row for timeslot={$result['row']['timeslot']}");
            $this->table(array_keys($result['row']), [$result['row']]);

            return Command::SUCCESS;
        } catch (RuntimeException $e) {
            $this->error($e->getMessage());

            return Command::FAILURE;
        }
    }
}
