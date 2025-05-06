<?php

namespace App\Jobs;

use App\Models\Runbox;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Throwable;

class RunboxJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    public int $tries = 3;

    public function __construct(public int $runboxId) {}

    public function handle(): void
    {
        $runbox = Runbox::find($this->runboxId);

        if (! $runbox) {
            Log::warning("Runbox #{$this->runboxId} not found");

            return;
        }

        try {
            $runbox->update(['phase' => 'running']);

            $response = Http::post(env('RUNBOX_URL')."/{$runbox->type}", $runbox->payload)->throw();
            $data = $response->json();

            $runbox->update([
                'outs' => $data['logs'] ?? $data['outputsList'] ?? null,
                'cpu' => $data['cpu'] ?? null,
                'mem' => $data['mem'] ?? null,
                'time' => $data['time'] ?? null,
                'phase' => 'succeeded',
            ]);

            Log::info("Runbox #{$runbox->id} processed successfully.");
        } catch (Throwable $e) {
            Log::error("Failed to process runbox #{$runbox->id}", [
                'error' => $e->getMessage(),
            ]);

            $runbox->update(['phase' => 'failed']);
        }
    }
}
