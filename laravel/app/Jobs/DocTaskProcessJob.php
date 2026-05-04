<?php

namespace App\Jobs;

use App\Services\DocTaskService;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Log;
use Throwable;

class DocTaskProcessJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    public int $tries = 1;

    public function handle(DocTaskService $service): void
    {
        try {
            $service->processTask();
        } catch (Throwable $e) {
            Log::error('DocTaskProcessJob failed', [
                'error' => $e->getMessage(),
            ]);

            throw $e;
        }
    }
}
