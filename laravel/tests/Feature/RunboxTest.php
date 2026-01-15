<?php

namespace Tests\Feature;

use App\Models\Runbox;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Tests\TestCase;

class RunboxTest extends TestCase
{
    public function test_runbox_get_nonexistent()
    {
        $response = $this->getJson('/api/runbox/nonexistent');
        $response->assertStatus(200);

        $data = $response->json();
        $this->assertEquals(['phase' => 'none'], $data);
    }

    public function test_runbox_post()
    {
        $tests = [
            [
                'data' => [
                    'hash' => 'lang_post_1',
                    'user_id' => 0,
                    'page_id' => 0,
                    'type' => 'lang',
                    'payload' => ['lang' => 'bash', 'files' => [['body' => 'echo lang_post_1']]],
                ],
                'wantCode' => 202,
                'wantOuts' => '["1lang_post_1"]',
            ],
            [
                'data' => [
                    'hash' => 'notebook_post_1',
                    'user_id' => 0,
                    'page_id' => 0,
                    'type' => 'notebook',
                    'payload' => ['lang' => 'python', 'sources' => ['msg = "notebook_post_1"', 'print(msg)']],
                ],
                'wantCode' => 202,
                'wantOuts' => '[[],[{"output_type":"stream","name":"stdout","text":["notebook_post_1\n"]}]]',
            ],
        ];

        config(['queue.default' => 'sync']);
        putenv('RUNBOX_URL=http://example.test');
        Http::fake(function ($request) {
            $url = $request->url();
            if (str_contains($url, '/lang')) {
                return Http::response([
                    'outputsList' => ['1lang_post_1'],
                    'cpu' => 1,
                    'mem' => 1,
                    'time' => 1,
                ], 200);
            }

            if (str_contains($url, '/notebook')) {
                return Http::response([
                    'outputsList' => [[], [[
                        'output_type' => 'stream',
                        'name' => 'stdout',
                        'text' => ["notebook_post_1\n"],
                    ]]],
                    'cpu' => 1,
                    'mem' => 1,
                    'time' => 1,
                ], 200);
            }

            return Http::response([], 404);
        });

        foreach ($tests as $t) {
            $data = $t['data'];
            $wantCode = $t['wantCode'];
            $wantOuts = $t['wantOuts'];
            $hash = $data['hash'];

            Runbox::where('hash', $hash)->delete();
            $this->withoutMiddleware();

            $response = $this->postJson('/api/runbox', $data);
            $response->assertStatus($wantCode);

            $this->waitForComplete($hash);

            $runbox = Runbox::where('hash', $hash)->first();
            $this->assertEquals('succeeded', $runbox->phase);
            $this->assertEquals($wantOuts, json_encode($runbox->outs));
            $this->assertNotEquals(0, $runbox->cpu);
            $this->assertNotEquals(0, $runbox->mem);
            $this->assertNotEquals(0, $runbox->time);
        }
    }

    private function waitForComplete($hash, $timeout = 20)
    {
        $startTime = time();
        while (true) {
            $row = Runbox::where('hash', $hash)->first();
            if ($row && $row->phase === 'succeeded') {
                return;
            }
            if ($row && $row->phase === 'failed') {
                $this->fail('Runbox failed');
            }
            Log::debug('Waiting', [
                'type' => $row?->type,
                'phase' => $row?->phase,
            ]);
            if ((time() - $startTime) > $timeout) {
                $this->fail('Timeout');
            }

            usleep(500000); // 0.5s
        }
    }
}
