<?php

namespace App\Console\Commands;

use App\Services\Stat\CollectMwHourlyService;
use Illuminate\Console\Command;
use RuntimeException;

class CollectMwHourlyCommand extends Command
{
    protected $signature = 'z:collect-mw-hourly
                            {--at= : Target UTC hour (YYYY-MM-DDTHH:00:00Z), default is current UTC hour}
                            {--debug : Print raw JSON payload}';
    protected $description = 'Collect MediaWiki hourly data';

    public function handle(CollectMwHourlyService $hourly): int
    {
        try {
            $result = $hourly->collect(trim((string) $this->option('at')));

            if ((bool) $this->option('debug')) {
                $this->line(json_encode($result['payload'], JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES) ?: '{}');
            }

            $this->info("Stored stat_mw_hourly row for timeslot={$result['timeslot_iso']}");
            $this->table(array_keys($result['row']), [$result['row']]);

            return Command::SUCCESS;
        } catch (RuntimeException $e) {
            $this->error($e->getMessage());

            return Command::FAILURE;
        }
    }
}
