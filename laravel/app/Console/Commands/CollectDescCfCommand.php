<?php

namespace App\Console\Commands;

use App\Services\Stat\CollectCfApiService;
use Illuminate\Console\Command;
use RuntimeException;

class CollectDescCfCommand extends Command
{
    protected $signature = 'z:collect-desc-cf';
    protected $description = 'List available Cloudflare stat datasets and fields';

    public function handle(CollectCfApiService $api): int
    {
        try {
            [$apiToken, $zoneId] = $api->resolveCredentials();

            $query = <<<'GRAPHQL'
query GetAvailableFields($zoneTag: string) {
  viewer {
    zones(filter: { zoneTag: $zoneTag }) {
      settings {
        httpRequests1hGroups {
          enabled
          availableFields
        }
        httpRequests1dGroups {
          enabled
          availableFields
        }
        httpRequestsAdaptiveGroups {
          enabled
          availableFields
        }
      }
    }
  }
}
GRAPHQL;

            $payload = $api->runGraphql($apiToken, $query, ['zoneTag' => $zoneId]);

            $settings = data_get($payload, 'data.viewer.zones.0.settings', []);
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
