<?php

use App\Support\StatWindow;
use Carbon\CarbonImmutable;

it('shows data up to two hours back before the hourly display cutoff', function () {
    $end = StatWindow::hourlyEnd(CarbonImmutable::parse('2026-03-16 16:07:00', 'UTC'));

    expect($end->format('Y-m-d H:i:s'))->toBe('2026-03-16 15:00:00');
});

it('shows data up to the previous hour at and after the hourly display cutoff', function () {
    $end = StatWindow::hourlyEnd(CarbonImmutable::parse('2026-03-16 16:10:00', 'UTC'));

    expect($end->format('Y-m-d H:i:s'))->toBe('2026-03-16 16:00:00');
});
