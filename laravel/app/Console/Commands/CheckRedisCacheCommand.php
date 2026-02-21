<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;
use Illuminate\Support\Facades\Cache;

class CheckRedisCacheCommand extends Command
{
    protected $signature = 'app:check-redis-cache';
    protected $description = 'Check redis cache';

    public function handle()
    {
        $ok = Cache::store('redis')->put('redis-cache-test', 'ok', 10);
        $this->info($ok ? '✅ Redis cache is working' : '❌ Failed to write to cache');
    }
}
