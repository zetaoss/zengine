<?php

namespace App\Console\Commands;

use Carbon\Carbon;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;

class QueueList extends Command
{
    protected $signature = 'queue:list';

    protected $description = 'List all queued jobs';

    public function handle()
    {
        $jobs = DB::table('jobs')->get();

        if ($jobs->isEmpty()) {
            $this->info('No jobs in the queue.');
        } else {
            $jobData = $jobs->map(function ($job) {
                $payload = json_decode($job->payload, true);
                $age = str_replace(' ago', '', Carbon::parse($job->created_at)->shortRelativeDiffForHumans());

                return [
                    'ID' => $job->id,
                    'Queue' => $job->queue,
                    'Name' => $payload['displayName'] ?? 'N/A',
                    'MaxTries' => $payload['maxTries'] ?? 'N/A',
                    'Attempts' => $job->attempts,
                    'Age' => $age,
                ];
            })->toArray();

            $columns = array_keys($jobData[0]);
            $this->table($columns, $jobData);
        }

        return Command::SUCCESS;
    }
}
