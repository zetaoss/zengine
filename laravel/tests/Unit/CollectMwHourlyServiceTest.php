<?php

use App\Services\Stat\CollectMwHourlyService;
use Carbon\CarbonImmutable;

afterEach(function () {
    CarbonImmutable::setTestNow();
});

it('defaults to the last completed utc hour', function () {
    CarbonImmutable::setTestNow(CarbonImmutable::parse('2026-03-16 16:55:00', 'UTC'));

    $service = new CollectMwHourlyService;
    $method = new ReflectionMethod($service, 'resolveTargetHour');
    $method->setAccessible(true);

    $result = $method->invoke($service, '');

    expect($result->format('Y-m-d H:i:s'))->toBe('2026-03-16 15:00:00');
});

it('normalizes an explicit utc hour input to the hour start', function () {
    $service = new CollectMwHourlyService;
    $method = new ReflectionMethod($service, 'resolveTargetHour');
    $method->setAccessible(true);

    $result = $method->invoke($service, '2026-03-16T15:37:24Z');

    expect($result->format('Y-m-d H:i:s'))->toBe('2026-03-16 15:00:00');
    expect($result->timezone->getName())->toBe('UTC');
});
