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
        $this->hash    = 'c398f8654f8ab6c44c4e8e9f6c438da7';
    }

    public function test_runbox_job_ok()
    {
        Runbox::where('hash', $this->hash)->delete();

        $runbox          = new Runbox();
        $runbox->type    = 'run';
        $runbox->state   = 0;
        $runbox->user_id = 0;
        $runbox->page_id = $this->page_id;
        $runbox->hash    = $this->hash;
        $runbox->payload = [
            "lang"  => "bash",
            "files" => [
                ["name" => "greet.txt", "body" => "hello world"],
                ["body" => "cat greet.txt"],
            ],
            "main"  => 1,
        ];
        $runbox->save();

        $job = new RunboxJob($runbox->id);
        $job->handle();

        $runbox->refresh();
        $this->assertEquals(3, $runbox->state);
        $this->assertEquals(["1hello world"], $runbox->logs);
        $this->assertNotEquals(0, $runbox->cpu);
        $this->assertNotEquals(0, $runbox->mem);
        $this->assertNotEquals(0, $runbox->time);
    }

    public function test_runbox_ok()
    {
        $response = $this->getJson("/api/runbox/{$this->page_id}/{$this->hash}");
        $response->assertStatus(200);

        $data = $response->json();
        $this->assertEquals(3, $data['state']);
        $this->assertEquals(0, $data['page_id']);
        $this->assertEquals(["1hello world"], $data['logs']);
        $this->assertNotEquals(0, $data['cpu']);
        $this->assertNotEquals(0, $data['mem']);
        $this->assertNotEquals(0, $data['time']);
    }

    public function test_runbox_notexist()
    {
        $response = $this->getJson('/api/runbox/0/nonexistent-hash');
        $response->assertStatus(200);

        $data = $response->json();
        $this->assertEquals(0, $data['state']);
    }
}
