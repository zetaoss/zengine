<?php

namespace App\Console\Commands;

use App\Services\DocTaskService;
use Illuminate\Console\Command;
use Throwable;

class ProcessDocTaskCommand extends Command
{
    protected $signature = 'z:process-doc-task';
    protected $description = 'Process the head docfac task, retrying the same task after its retry delay';

    public function handle(DocTaskService $service): int
    {
        try {
            $task = $service->processTask();
        } catch (Throwable $e) {
            $this->error($e->getMessage());

            return Command::FAILURE;
        }

        if (! $task) {
            $this->info('No eligible doc task to process.');

            return Command::SUCCESS;
        }

        $this->info("Processed doc task #{$task->id}: {$task->title}");

        return Command::SUCCESS;
    }
}
