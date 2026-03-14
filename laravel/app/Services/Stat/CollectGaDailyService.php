<?php

namespace App\Services\Stat;

use App\Models\StatGaDaily;

class CollectGaDailyService
{
    public function __construct(
        private readonly CollectGaApiService $api,
        private readonly CollectGaPersistService $persist,
    ) {}

    public function collect(int $days, bool $debug, callable $debugWriter): array
    {
        [$propertyId, $clientEmail, $privateKey, $timezone, $tokenUri] = $this->api->resolveCredentials();
        [$sinceLocal, $untilLocal] = $this->api->propertyDateRangeForDays($days, $timezone);

        $requestBody = [
            'dateRanges' => [[
                'startDate' => $sinceLocal->toDateString(),
                'endDate' => $untilLocal->toDateString(),
            ]],
            'dimensions' => [['name' => 'date']],
            'metrics' => [
                ['name' => 'activeUsers'],
                ['name' => 'screenPageViews'],
                ['name' => 'sessions'],
            ],
            'orderBys' => [[
                'dimension' => ['dimensionName' => 'date'],
            ]],
            'keepEmptyRows' => true,
            'limit' => 10000,
        ];

        if ($debug) {
            $debugWriter(
                'Debug request for Google Analytics daily runReport:',
                $this->api->encodeDebugJson($requestBody)
            );
        }

        $accessToken = $this->api->fetchAccessToken($clientEmail, $privateKey, $tokenUri);
        $payload = $this->api->runReport($accessToken, $propertyId, $requestBody);

        if ($debug) {
            $debugWriter(
                'Debug payload for Google Analytics daily runReport:',
                $this->api->encodeDebugJson($payload)
            );
        }

        $rows = [];
        foreach ((array) data_get($payload, 'rows', []) as $row) {
            $dateRaw = (string) data_get($row, 'dimensionValues.0.value', '');
            $timeslot = $this->api->normalizeDateDimension($dateRaw);
            if ($timeslot === null) {
                continue;
            }

            $rows[] = [
                'timeslot' => $timeslot,
                'active_users' => $this->metricValue($row, 0),
                'screen_page_views' => $this->metricValue($row, 1),
                'sessions' => $this->metricValue($row, 2),
            ];
        }

        return [
            'propertyId' => $propertyId,
            'timezone' => $timezone,
            'sinceLocal' => $sinceLocal,
            'untilLocal' => $untilLocal,
            'timeslots' => array_values(array_map(static fn (array $row): string => (string) $row['timeslot'], $rows)),
            'db' => $this->persist->persistRows(StatGaDaily::class, $rows),
        ];
    }

    private function metricValue(array $row, int $index): int
    {
        $value = data_get($row, "metricValues.{$index}.value");

        return is_numeric($value) ? (int) $value : 0;
    }
}
