<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;

class CollectHourlyCommand extends Command
{
    protected $signature = 'z:collect-hourly
                            {--days=2 : Number of trailing days to use for CF, GA, and GSC hourly collection}
                            {--debug : Print raw debug output from each hourly collector}';
    protected $description = 'Collect all hourly stat sources';

    public function handle(): int
    {
        $days = (int) $this->option('days');
        $debug = (bool) $this->option('debug');

        $commands = [
            'z:collect-cf-hourly' => ['--days' => $days, '--debug' => $debug],
            'z:collect-ga-hourly' => ['--days' => $days, '--debug' => $debug],
            'z:collect-gsc-hourly' => ['--days' => $days, '--debug' => $debug],
            'z:collect-mw-hourly' => ['--debug' => $debug],
        ];

        foreach ($commands as $name => $arguments) {
            $this->newLine();
            $this->info("Running {$name}...");

            $exitCode = $this->call($name, $arguments);

            if ($exitCode !== self::SUCCESS) {
                $this->error("Hourly collection stopped because {$name} failed.");

                return $exitCode;
            }
        }

        $this->newLine();
        $this->info('All hourly collectors completed.');

        return self::SUCCESS;
    }
}
