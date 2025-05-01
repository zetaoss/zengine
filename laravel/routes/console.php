<?php

use Illuminate\Support\Facades\Schedule;

Schedule::useCache('redis');
Schedule::command('app:common-report-timeout')->everyMinute()->onOneServer();
