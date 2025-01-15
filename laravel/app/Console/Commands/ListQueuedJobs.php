<?php
namespace App\Console\Commands;

use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;

class ListQueuedJobs extends Command
{
    protected $signature   = 'queue:list';
    protected $description = 'List all queued jobs';

    public function handle()
    {
        $jobs = DB::table('jobs')->get();

        if ($jobs->isEmpty()) {
            $this->info('No jobs in the queue.');
        } else {
            foreach ($jobs as $job) {
                $this->line("ID: {$job->id}, Queue: {$job->queue}, Payload: {$job->payload}");
            }
        }

        return Command::SUCCESS;
    }
}
