<?php
namespace Tests\Unit;

use App\Jobs\RunboxJob;
use App\Models\Runbox;
use Illuminate\Support\Str;
use Tests\TestCase;

class RunboxJobTest extends TestCase
{
    public function testOk()
    {
        $runbox          = new Runbox();
        $runbox->type    = 'run';
        $runbox->state   = 0;
        $runbox->user_id = 0;
        $runbox->page_id = 0;
        $runbox->hash    = md5(Str::uuid());
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
}
