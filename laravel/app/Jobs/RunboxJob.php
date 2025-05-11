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
            $runbox->phase = 'running';
            $runbox->save();

            $type = $runbox->type;
            $response = Http::post(env('RUNBOX_URL')."/$type", $runbox->payload)->throw();
            $result = $response->json();

            if ($type === 'notebook') {
                $runbox->outs = $result['outputsList'] ?? [];
            } else {
                $runbox->outs = [
                    'logs' => $result['logs'] ?? [],
                    'images' => $result['images'] ?? [],
                ];
            }

            $runbox->cpu = $result['cpu'] ?? null;
            $runbox->mem = $result['mem'] ?? null;
            $runbox->time = $result['time'] ?? null;
            $runbox->phase = 'succeeded';
            $runbox->save();

        } catch (Throwable $e) {
            Log::error("RunboxJob failed for ID {$runbox->id}: {$e->getMessage()}");
            $runbox->phase = 'failed';
            $runbox->save();
        }
    }
}
