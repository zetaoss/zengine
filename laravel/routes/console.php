<?php

use Illuminate\Support\Facades\Schedule;

Schedule::command('app:cleanup-not-matches')->daily();
Schedule::command('app:cf-analytics-hourly-collect')->hourlyAt(10)->withoutOverlapping();
Schedule::command('app:cf-analytics-daily-collect')->hourlyAt(10)->withoutOverlapping();
Schedule::command('app:mw-statistics-collect')->cron('0 10,12,14 * * *')->withoutOverlapping();
