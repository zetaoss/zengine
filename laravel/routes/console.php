<?php

use Illuminate\Support\Facades\Schedule;

Schedule::command('z:collect-daily-cf')->cron('10 10,12,14 * * *')->withoutOverlapping();
Schedule::command('z:collect-daily-mw')->cron('10 10,12,14 * * *')->withoutOverlapping();
Schedule::command('z:collect-hourly-cf')->hourlyAt(5)->withoutOverlapping();
Schedule::command('z:collect-hourly-mw')->hourlyAt(10)->withoutOverlapping();
Schedule::command('z:cleanup-not-matches')->daily()->withoutOverlapping();
Schedule::command('z:refresh-write-request')->hourlyAt(15)->withoutOverlapping();
