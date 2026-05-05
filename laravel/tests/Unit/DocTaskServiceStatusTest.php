<?php

use App\Services\DocTaskService;
use App\Services\LLMService;
use Illuminate\Support\Carbon;
use Mockery;

it('keeps next_run_at stable from the recorded state change time', function () {
    $service = new DocTaskService(Mockery::mock(LLMService::class));
    $method = new ReflectionMethod(DocTaskService::class, 'nextRunAt');
    $method->setAccessible(true);

    Carbon::setTestNow('2026-05-06T02:52:03+09:00');
    $first = $method->invoke($service, [
        'state_changed_at' => '2026-05-06T02:52:00+09:00',
    ], 240);

    Carbon::setTestNow('2026-05-06T02:53:00+09:00');
    $second = $method->invoke($service, [
        'state_changed_at' => '2026-05-06T02:52:00+09:00',
    ], 240);

    expect($first)->toBe('2026-05-06T02:56:00+09:00');
    expect($second)->toBe('2026-05-06T02:56:00+09:00');

    Carbon::setTestNow();
});
