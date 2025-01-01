<?php
namespace App\Jobs;

use App\Models\Box;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Http;
use Throwable;

class BoxJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;
    private $box;

    public function __construct($boxID)
    {
        $this->box = Box::find($boxID);
    }

    public function handle()
    {
        $this->box->step = 2;
        $this->box->save();

        $data = [
            "lang" => "bash",
            "files" => [
                ["name" => "greet.txt", "body" => "hello"],
                ["name" => "", "body" => "cat greet.txt"],
            ],
            "main" => 1,
        ];
        $url = "runbox:80/run/lang";
        $response = Http::post($url, $data);
        if ($response->successful()) {
            return $response->json();
        } else {
            return response()->json(['error' => 'failed to send data'], $response->status());
        }
        $api = $this->box->api;
        $data = [
            'lang' => $this->box->lang,
            'source' => $this->box->source,
        ];
        $body = json_encode($data);
        try {
            $response = Http::post($url, $data);
            if ($response->successful()) {
                return $response->json();
            } else {
                return response()->json(['error' => 'failed to send data'], $response->status());
            }

            $resp = Http::post($url, $data);
            $arr = json_decode($resp->body(), true);
            if (array_key_exists('errorcode', $arr)) {
                var_dump('Error...', $arr);
                $this->box->metadata = [$arr['errorcode']];
                $this->box->outs = [$arr['msg']];
                $this->box->step = 9;
                $this->box->save();
            } else {
                $this->box->cpu = $arr['cpu'];
                $this->box->mem = $arr['mem'];
                $this->box->time = $arr['time'];
                $this->box->code = $arr['code'];
                $this->box->metadata = $arr['metadata'];
                $this->box->outs = $arr['outs'];
                $this->box->step = 3;
                $this->box->save();
                var_dump('Saved. (step=3)');
            }
        } catch (Exception $e) {
            $this->box->step = 9;
            $this->box->save();
            var_dump('error=', $e->getMessage());
        }
    }

    public function failed(Throwable $exception)
    {
        var_dump($exception);
        $this->box->step = 9;
        $this->box->save();
    }

}
