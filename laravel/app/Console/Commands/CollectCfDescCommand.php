<?php

namespace App\Console\Commands;

use App\Services\Stat\CollectCfDescService;
use Illuminate\Console\Command;
use RuntimeException;

class CollectCfDescCommand extends Command
{
    protected $signature = 'z:collect-cf-desc';
    protected $description = 'Describe Cloudflare analytics available data fields';

    public function handle(CollectCfDescService $desc): int
    {
        try {
            $result = $desc->collect();
            $settings = $result['settings'];
            if (empty($settings)) {
                $this->warn('No settings data found for this zone.');

                return Command::SUCCESS;
            }

            foreach (['httpRequests1dGroups', 'httpRequests1hGroups', 'httpRequestsAdaptiveGroups'] as $dataset) {
                $enabled = (bool) data_get($settings, "{$dataset}.enabled", false);
                $availableFields = (array) data_get($settings, "{$dataset}.availableFields", []);

                $this->info("Dataset: {$dataset} (enabled: ".($enabled ? 'yes' : 'no').')');
                if (empty($availableFields)) {
                    $this->line('- no available fields');

                    continue;
                }

                foreach ($availableFields as $field) {
                    $this->line('- '.$field);
                }
            }

            return Command::SUCCESS;
        } catch (RuntimeException $e) {
            $this->error($e->getMessage());

            return Command::FAILURE;
        }
    }
}
