<?php

use Illuminate\Support\Facades\Schedule;

Schedule::command('app:cleanup-not-matches')->daily();
