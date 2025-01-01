<?php

use App\Services\RunService;

function assertEqualResult(array $want, array $got): void
{
    expect($got['cpu'])->toBeGreaterThan($want['cpu'] / 10);
    expect($got['mem'])->toBeGreaterThan($want['mem'] / 10);
    expect($got['time'])->toBeGreaterThan($want['time'] / 10);
    expect($got['cpu'])->toBeLessThan($want['cpu'] * 100);
    expect($got['mem'])->toBeLessThan($want['mem'] * 100);
    expect($got['time'])->toBeLessThan($want['time'] * 100);
    $want['cpu'] = $got['cpu'];
    $want['mem'] = $got['mem'];
    $want['time'] = $got['time'];

    expect($got)->toBe($want);
}

it('lang', function () {
    $input = [
        "lang" => "bash",
        "files" => [
            ["name" => "greet.txt", "body" => "hello"],
            ["name" => "", "body" => "cat greet.txt"],
        ],
        "main" => 1,
    ];
    $want = [
        'logs' => ['1hello'],
        'cpu' => 37384,
        'mem' => 360,
        'time' => 75,
    ];
    [$got, $err] = RunService::lang($input);

    $this->assertNull($err);
    assertEqualResult($want, $got);
});
