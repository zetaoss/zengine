<?php

namespace App\Support;

use Carbon\CarbonImmutable;
use Carbon\CarbonInterface;

class StatWindow
{
    public static function hourlyEnd(?CarbonInterface $now = null, int $readyMinute = 10): CarbonImmutable
    {
        $current = $now
            ? CarbonImmutable::instance($now)
            : CarbonImmutable::now('UTC');

        $hoursBack = $current->minute < $readyMinute ? 1 : 0;

        return $current->startOfHour()->subHours($hoursBack);
    }

    public static function dailyEnd(?CarbonInterface $now = null): CarbonImmutable
    {
        $current = $now
            ? CarbonImmutable::instance($now)
            : CarbonImmutable::now('UTC');

        return $current->startOfDay();
    }
}
