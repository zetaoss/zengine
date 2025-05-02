<?php

namespace App\Console\Commands;

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

        $now = new \DateTime;

        $rows = $jobs->map(function ($job) use ($now) {
            $created = new \DateTime($job->created_at ?? 'now');
            $age = $this->formatElapsedTime($created, $now);

            $payload = json_decode($job->payload, true) ?? [];

            return [
                'ID' => $job->id,
                'Queue' => $job->queue,
                'Name' => $payload['displayName'] ?? 'N/A',
                'Tries' => ($payload['maxTries'] ?? '∞').' / '.$job->attempts,
                'Age' => $age,
            ];
        })->toArray();

        $this->table(array_keys($rows[0]), $rows);

        return Command::SUCCESS;
    }

    protected function formatElapsedTime(\DateTime $from, \DateTime $to): string
    {
        $elapsedSeconds = max(0, $to->getTimestamp() - $from->getTimestamp());

        $hours = intdiv($elapsedSeconds, 3600);
        $minutes = intdiv($elapsedSeconds % 3600, 60);
        $seconds = $elapsedSeconds % 60;

        return sprintf('%d:%02d:%02d', $hours, $minutes, $seconds);
    }
}
