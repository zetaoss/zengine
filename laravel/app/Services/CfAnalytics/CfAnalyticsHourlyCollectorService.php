<?php

namespace App\Services\CfAnalytics;

use App\Models\CfAnalyticsHourly;
use Carbon\CarbonImmutable;
use RuntimeException;

class CfAnalyticsHourlyCollectorService
{
    private const MAX_DAYS = 3;

    public function __construct(
        private readonly CfAnalyticsApiService $api,
    ) {}

    public function collect(int $days, bool $debug, callable $debugWriter): array
    {
        if ($days > self::MAX_DAYS) {
            throw new RuntimeException(
                '--days for app:cf-analytics-hourly-collect must be 1..3 (Cloudflare 1h dataset retention is about 262800 seconds).'
            );
        }

        [$apiToken, $zoneId] = $this->api->resolveCredentials();
        [$sinceUtc, $untilUtc] = $this->api->utcHourAnchoredRangeForDays($days);

        $rowsByTimeslot = [];
        foreach ($this->api->utcDayWindows($sinceUtc, $untilUtc) as $window) {
            $variables = [
                'zoneTag' => $zoneId,
                'since' => $window['start_utc']->toIso8601String(),
                'until' => $window['end_utc']->toIso8601String(),
            ];
            $query = $this->hourlyQuery();

            if ($debug) {
                $this->writeDebugRequest(
                    $debugWriter,
                    "Debug request for httpRequests1hGroups {$window['date_utc']}:",
                    $query,
                    $variables,
                    [
                        'date_utc' => $window['date_utc'],
                        'range_utc' => [
                            'since' => $window['start_utc']->toIso8601String(),
                            'until' => $window['end_utc']->toIso8601String(),
                        ],
                    ]
                );
            }

            $payload = $this->api->runGraphql($apiToken, $query, $variables);
            if ($debug) {
                $debugWriter(
                    "Debug payload for httpRequests1hGroups {$window['date_utc']}:",
                    $this->api->encodeDebugJson($payload)
                );
            }

            $groups = (array) data_get($payload, 'data.viewer.zones.0.zones', []);
            foreach ($groups as $group) {
                $timeslotRaw = (string) data_get($group, 'dimensions.timeslot', '');
                if ($timeslotRaw === '') {
                    continue;
                }
                $timeslot = CarbonImmutable::parse($timeslotRaw)->utc()->format('Y-m-d H:i:s');

                $rowsByTimeslot[$timeslot] = [
                    'uniq_uniques' => $this->toTextValue(data_get($group, 'uniq.uniques', 0)),
                    'sum_requests' => $this->toTextValue(data_get($group, 'sum.requests', 0)),
                    'sum_pageViews' => $this->toTextValue(data_get($group, 'sum.pageViews', 0)),
                    'sum_bytes' => $this->toTextValue(data_get($group, 'sum.bytes', 0)),
                    'sum_cachedBytes' => $this->toTextValue(data_get($group, 'sum.cachedBytes', 0)),
                    'sum_cachedRequests' => $this->toTextValue(data_get($group, 'sum.cachedRequests', 0)),
                    'sum_encryptedBytes' => $this->toTextValue(data_get($group, 'sum.encryptedBytes', 0)),
                    'sum_encryptedRequests' => $this->toTextValue(data_get($group, 'sum.encryptedRequests', 0)),
                    'sum_threats' => $this->toTextValue(data_get($group, 'sum.threats', 0)),
                    'sum_browserMap' => $this->toTextValue((array) data_get($group, 'sum.browserMap', [])),
                    'sum_contentTypeMap' => $this->toTextValue((array) data_get($group, 'sum.contentTypeMap', [])),
                    'sum_clientSSLMap' => $this->toTextValue((array) data_get($group, 'sum.clientSSLMap', [])),
                    'sum_countryMap' => $this->toTextValue((array) data_get($group, 'sum.countryMap', [])),
                    'sum_ipClassMap' => $this->toTextValue((array) data_get($group, 'sum.ipClassMap', [])),
                    'sum_responseStatusMap' => $this->toTextValue((array) data_get($group, 'sum.responseStatusMap', [])),
                    'sum_threatPathingMap' => $this->toTextValue((array) data_get($group, 'sum.threatPathingMap', [])),
                ];
            }
        }

        $rows = [];
        foreach ($rowsByTimeslot as $timeslot => $metrics) {
            foreach ($metrics as $name => $value) {
                $rows[] = [
                    'timeslot' => $timeslot,
                    'name' => $name,
                    'value' => $value,
                ];
            }
        }

        return [
            'zoneId' => $zoneId,
            'sinceUtc' => $sinceUtc,
            'untilUtc' => $untilUtc,
            'timeslots' => array_values(array_keys($rowsByTimeslot)),
            'db' => $this->persistRows($rows),
        ];
    }

    private function persistRows(array $rows): array
    {
        if (empty($rows)) {
            return ['inserted' => 0, 'updated' => 0, 'skipped' => 0];
        }

        $timeslots = array_values(array_unique(array_map(fn ($row) => (string) $row['timeslot'], $rows)));
        $names = array_values(array_unique(array_map(fn ($row) => (string) $row['name'], $rows)));

        $existing = CfAnalyticsHourly::query()
            ->toBase()
            ->select(['timeslot', 'name', 'value'])
            ->whereIn('timeslot', $timeslots)
            ->whereIn('name', $names)
            ->get();

        $existingValues = [];
        foreach ($existing as $item) {
            $existingValues[$this->metricKey((string) $item->timeslot, (string) $item->name)] = (string) $item->value;
        }

        $upsertRows = [];
        $inserted = 0;
        $updated = 0;
        $skipped = 0;
        foreach ($rows as $row) {
            $key = $this->metricKey((string) $row['timeslot'], (string) $row['name']);
            $newValue = (string) $row['value'];
            $oldValue = $existingValues[$key] ?? null;

            if ($oldValue === null) {
                $inserted++;
                $upsertRows[] = $row;

                continue;
            }

            if ($oldValue !== $newValue) {
                $updated++;
                $upsertRows[] = $row;
            } else {
                $skipped++;
            }
        }

        if (! empty($upsertRows)) {
            CfAnalyticsHourly::query()->upsert($upsertRows, ['timeslot', 'name'], ['value']);
        }

        return ['inserted' => $inserted, 'updated' => $updated, 'skipped' => $skipped];
    }

    private function metricKey(string $timeslot, string $name): string
    {
        return $timeslot.'|'.$name;
    }

    private function toTextValue(mixed $value): string
    {
        if (is_array($value)) {
            $json = json_encode($value, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);

            return $json === false ? '[]' : $json;
        }

        if ($value === null) {
            return '';
        }

        return (string) $value;
    }

    private function writeDebugRequest(
        callable $debugWriter,
        string $label,
        string $query,
        array $variables,
        array $rangeMeta = []
    ): void {
        $debugWriter($label, $this->api->encodeDebugJson([
            'range' => $rangeMeta,
            'variables' => $variables,
            'query' => $query,
        ]));
    }

    private function hourlyQuery(): string
    {
        return <<<'GRAPHQL'
query GetZoneAnalytics($zoneTag: string, $since: string, $until: string) {
  viewer {
    zones(filter: { zoneTag: $zoneTag }) {
      zones: httpRequests1hGroups(
        orderBy: [datetime_ASC]
        limit: 10000
        filter: { datetime_geq: $since, datetime_lt: $until }
      ) {
        dimensions {
          timeslot: datetime
        }
        uniq {
          uniques
        }
        sum {
          browserMap {
            pageViews
            key: uaBrowserFamily
          }
          bytes
          cachedBytes
          cachedRequests
          contentTypeMap {
            bytes
            requests
            key: edgeResponseContentTypeName
          }
          clientSSLMap {
            requests
            key: clientSSLProtocol
          }
          countryMap {
            bytes
            requests
            threats
            key: clientCountryName
          }
          encryptedBytes
          encryptedRequests
          ipClassMap {
            requests
            key: ipType
          }
          pageViews
          requests
          responseStatusMap {
            requests
            key: edgeResponseStatus
          }
          threats
          threatPathingMap {
            requests
            key: threatPathingName
          }
        }
      }
    }
  }
}
GRAPHQL;
    }
}
