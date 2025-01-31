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
    private $runbox;

    public function __construct($runboxID)
    {
        $this->runbox = Runbox::find($runboxID);
        if (! $this->runbox) {
            throw new \Exception("Runbox with ID {$runboxID} not found.");
        }
    }

    public function handle()
    {
        try {
            $this->runbox->step = 2;
            $this->runbox->save();

            $resp = Http::post(getenv('RUNBOX_URL') . '/lang', $this->runbox->payload)->throw();
            $arr  = json_decode($resp->body(), true);

            $this->runbox->logs = $arr['logs'];
            $this->runbox->cpu  = $arr['cpu'];
            $this->runbox->mem  = $arr['mem'];
            $this->runbox->time = $arr['time'];
            $this->runbox->step = 3;
            $this->runbox->save();

        } catch (Throwable $e) {
            $this->failed($e);
            throw $e;
        }
    }

    public function failed(Throwable $exception)
    {
        Log::error("RunboxJob failed for Runbox ID: {$this->runbox->id}");
        $this->runbox->step = 9;
        $this->runbox->save();
    }
}
