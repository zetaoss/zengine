<?php

use Illuminate\Support\Facades\Schedule;

Schedule::command('z:cf-analytics-daily-collect')->hourlyAt(5)->withoutOverlapping();
Schedule::command('z:cf-analytics-hourly-collect')->hourlyAt(10)->withoutOverlapping();
Schedule::command('z:cleanup-not-matches')->daily()->withoutOverlapping();
Schedule::command('z:mw-statistics-collect')->cron('0 10,12,14 * * *')->withoutOverlapping();
Schedule::command('z:refresh-write-request')->hourlyAt(15)->withoutOverlapping();
