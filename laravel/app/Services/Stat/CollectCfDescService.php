<?php

namespace App\Services\Stat;

class CollectCfDescService
{
    public function __construct(
        private readonly CollectCfApiService $api,
    ) {}

    public function collect(): array
    {
        [$apiToken, $zoneId] = $this->api->resolveCredentials();

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

        $payload = $this->api->runGraphql($apiToken, $query, ['zoneTag' => $zoneId]);

        return [
            'zoneId' => $zoneId,
            'payload' => $payload,
            'settings' => data_get($payload, 'data.viewer.zones.0.settings', []),
        ];
    }
}
