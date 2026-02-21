<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;

class CleanupNotMatchesCommand extends Command
{
    protected $signature = 'app:cleanup-not-matches {max_hit=1}';
    protected $description = 'Delete low-signal not_matches rows (hit <= max_hit)';

    public function handle(): int
    {
        $maxHit = (int) $this->argument('max_hit');
        if ($maxHit < 0) {
            $this->error('max_hit must be a non-negative integer.');

            return Command::FAILURE;
        }

        $cutoff = now()->subYear();

        $deleted = DB::table('not_matches')
            ->where('hit', '<=', $maxHit)
            ->where('updated_at', '<=', $cutoff)
            ->delete();

        $this->info("{$deleted} row(s) deleted from not_matches (hit <= {$maxHit}, updated_at <= {$cutoff}).");

        return Command::SUCCESS;
    }
}
