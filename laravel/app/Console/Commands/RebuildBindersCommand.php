<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;

class RebuildBindersCommand extends Command
{
    protected $signature = 'z:binder-rebuild-all';
    protected $description = 'Run the MediaWiki binder rebuild maintenance script';

    public function handle(): int
    {
        $command = PHP_BINARY.' '.getenv('MW_INSTALL_PATH').'/extensions/ZetaExtension/maintenance/RebuildBinders.php';

        passthru($command, $exitCode);

        return $exitCode === 0 ? Command::SUCCESS : Command::FAILURE;
    }
}
