<?php

use Illuminate\Support\Facades\Schedule;

Schedule::command('z:collect-cf-daily --days=10')->hourlyAt(5)->withoutOverlapping();
Schedule::command('z:collect-ga-daily --days=10')->hourlyAt(5)->withoutOverlapping();
Schedule::command('z:collect-gsc-daily --days=10')->hourlyAt(5)->withoutOverlapping();
Schedule::command('z:collect-mw-daily')->hourlyAt(5)->withoutOverlapping();
Schedule::command('z:collect-cf-hourly')->hourlyAt(5)->withoutOverlapping();
Schedule::command('z:collect-ga-hourly')->hourlyAt(5)->withoutOverlapping();
Schedule::command('z:collect-gsc-hourly')->hourlyAt(5)->withoutOverlapping();
Schedule::command('z:collect-mw-hourly')->hourlyAt(5)->withoutOverlapping();
Schedule::command('z:refresh-write-request')->hourlyAt(15)->withoutOverlapping();
Schedule::command('z:cleanup-not-matches')->daily()->withoutOverlapping();
Schedule::command('z:process-doc-task')->everyMinute()->withoutOverlapping();
