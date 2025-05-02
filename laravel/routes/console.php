<?php

use Illuminate\Support\Facades\Schedule;

Schedule::command('app:common-report-timeout')
    ->everyMinute()
    ->appendOutputTo('/proc/1/fd/1');
