<?php
namespace Tests\Unit;

use App\Services\RunService;
use Tests\TestCase;

class RunboxTest extends TestCase
{
    public function testOk()
    {
        $data = [
            "lang"  => "bash",
            "files" => [
                ["name" => "greet.txt", "body" => "hello"],
                ["body" => "cat greet.txt"],
            ],
            "main"  => 1,
        ];

        [$json, $err] = RunService::lang($data);

        $this->assertNull($err);
        $this->assertNotNull($json);
        $this->assertArrayHasKey('logs', $json);
        $this->assertEquals('1hello', trim($json['logs'][0]));
        $this->assertArrayHasKey('cpu', $json);
        $this->assertArrayHasKey('mem', $json);
        $this->assertArrayHasKey('time', $json);
    }

    public function testError()
    {
        $data = [
            "lang"  => "invalid-lang",
            "files" => [
                ["name" => "greet.txt", "body" => "hello"],
                ["body" => "cat greet.txt"],
            ],
            "main"  => 1,
        ];

        [$json, $err] = RunService::lang($data);

        $this->assertNotNull($err);
        $this->assertNull($json);
        $this->assertStringContainsString('unsuccessful: 400', $err);
    }
}
