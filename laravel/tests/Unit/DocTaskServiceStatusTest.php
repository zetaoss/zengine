<?php

use App\Services\DocTaskService;
use App\Services\LLMService;
use Illuminate\Support\Carbon;

it('computes next_run_at from now regardless of state_changed_at', function () {
    $service = new DocTaskService(Mockery::mock(LLMService::class));
    $method = new ReflectionMethod(DocTaskService::class, 'nextRunAt');
    $method->setAccessible(true);

    Carbon::setTestNow('2026-05-06T02:52:03+09:00');
    $expectedFirst = now()->addSeconds(240)->toIso8601String();
    $first = $method->invoke($service, [
        'state_changed_at' => '2026-05-06T02:52:00+09:00',
    ], 240);

    Carbon::setTestNow('2026-05-06T02:53:00+09:00');
    $expectedSecond = now()->addSeconds(240)->toIso8601String();
    $second = $method->invoke($service, [
        'state_changed_at' => '2026-05-06T02:52:00+09:00',
    ], 240);

    expect($first)->toBe($expectedFirst);
    expect($second)->toBe($expectedSecond);
    expect($first)->not->toBe($second);

    Carbon::setTestNow();
});
