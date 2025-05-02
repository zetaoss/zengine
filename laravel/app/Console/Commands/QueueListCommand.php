<?php

namespace App\Console\Commands;

use Carbon\Carbon;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;

class QueueListCommand extends Command
{
    protected $signature = 'app:queue-list';

    protected $description = 'List all pending jobs in the queue';

    public function handle(): int
    {
        $jobs = DB::table('jobs')->orderByDesc('id')->get();

        if ($jobs->isEmpty()) {
            $this->info('✅ No jobs in the queue.');

            return Command::SUCCESS;
        }

        $rows = $jobs->map(function ($job) {
            $payload = json_decode($job->payload, true) ?? [];

            return [
                'ID' => $job->id,
                'Queue' => $job->queue,
                'Name' => $payload['displayName'] ?? 'N/A',
                'Tries' => ($payload['maxTries'] ?? '∞').' / '.$job->attempts,
                'Age' => Carbon::parse($job->created_at)->diffForHumans(now(), true), // e.g., "2m"
            ];
        })->toArray();

        $this->table(array_keys($rows[0]), $rows);

        return Command::SUCCESS;
    }
}
