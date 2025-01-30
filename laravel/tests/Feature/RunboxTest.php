<?php
namespace Tests\Feature;

use App\Jobs\RunboxJob;
use App\Models\Runbox;
use Tests\TestCase;

class RunboxTest extends TestCase
{
    protected int $page_id;
    protected string $hash;

    protected function setUp(): void
    {
        parent::setUp();

        $this->page_id = 0;
        $this->hash    = 'test1234';
    }

    public function test_runbox_job_ok()
    {
        Runbox::where('hash', $this->hash)->delete();

        $runbox          = new Runbox();
        $runbox->type    = 'run';
        $runbox->step    = 0;
        $runbox->user_id = 0;
        $runbox->page_id = $this->page_id;
        $runbox->hash    = $this->hash;
        $runbox->payload = [
            "lang"  => "bash",
            "files" => [
                ["name" => "greet.txt", "body" => "hello test"],
                ["body" => "cat greet.txt"],
            ],
            "main"  => 1,
        ];
        $runbox->save();

        $job = new RunboxJob($runbox->id);
        $job->handle();

        $runbox->refresh();
        $this->assertEquals(3, $runbox->step);
        $this->assertEquals(["1hello test"], $runbox->logs);
        $this->assertNotEquals(0, $runbox->cpu);
        $this->assertNotEquals(0, $runbox->mem);
        $this->assertNotEquals(0, $runbox->time);
    }

    public function test_runbox_get0()
    {
        $response = $this->getJson('/api/runbox/0/nonexistent-hash');
        $response->assertStatus(200);

        $data = $response->json();
        $this->assertEquals(0, $data['step']);
    }

    public function test_runbox_get3()
    {
        $response = $this->getJson("/api/runbox/{$this->page_id}/{$this->hash}");
        $response->assertStatus(200);

        $data = $response->json();
        $this->assertEquals(3, $data['step']);
        $this->assertEquals(0, $data['page_id']);
        $this->assertEquals(["1hello test"], $data['logs']);
        $this->assertNotEquals(0, $data['cpu']);
        $this->assertNotEquals(0, $data['mem']);
        $this->assertNotEquals(0, $data['time']);
    }

    public function test_runbox_post()
    {
        $page_id = 1;
        $hash    = 'post1234';
        $payload = [
            'lang'  => 'bash',
            'files' => [['body' => 'echo hello post']],
        ];

        Runbox::where('hash', $hash)->delete();
        $this->withoutMiddleware();
        $response = $this->postJson("/api/runbox/{$page_id}/{$hash}", $payload);
        $response->assertStatus(202);

        $this->waitForRunboxComplete($hash);

        $runbox = Runbox::where('hash', $hash)->first();
        $this->assertEquals(3, $runbox->step);
        $this->assertEquals(["1hello post"], $runbox->logs);
        $this->assertNotEquals(0, $runbox->cpu);
        $this->assertNotEquals(0, $runbox->mem);
        $this->assertNotEquals(0, $runbox->time);
    }

    private function waitForRunboxComplete($hash, $timeout = 20)
    {
        $startTime = time();
        while (true) {
            $row = Runbox::where('hash', $hash)->first();
            if ($row && $row->step > 2) {
                return;
            }

            if ((time() - $startTime) > $timeout) {
                $this->fail("Timeout");
            }

            usleep(500000); // 0.5s
        }
    }
}
