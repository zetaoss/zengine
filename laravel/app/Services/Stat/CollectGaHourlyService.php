<?php

namespace App\Services\Stat;

use App\Models\StatHourlyGa;

class CollectGaHourlyService
{
    public function __construct(
        private readonly CollectGaApiService $api,
    ) {}

    public function collect(int $days, bool $debug, callable $debugWriter): array
    {
        [$propertyId, $clientEmail, $privateKey, $timezone, $tokenUri] = $this->api->resolveCredentials();
        [$sinceLocal, $untilLocal] = $this->api->propertyHourAnchoredRangeForDays($days, $timezone);

        $requestBody = [
            'dateRanges' => [[
                'startDate' => $sinceLocal->toDateString(),
                'endDate' => $untilLocal->subHour()->toDateString(),
            ]],
            'dimensions' => [['name' => 'dateHour']],
            'metrics' => [
                ['name' => 'activeUsers'],
                ['name' => 'screenPageViews'],
                ['name' => 'sessions'],
            ],
            'orderBys' => [[
                'dimension' => ['dimensionName' => 'dateHour'],
            ]],
            'keepEmptyRows' => true,
            'limit' => 10000,
        ];

        if ($debug) {
            $debugWriter(
                'Debug request for Google Analytics hourly runReport:',
                $this->api->encodeDebugJson($requestBody)
            );
        }

        $accessToken = $this->api->fetchAccessToken($clientEmail, $privateKey, $tokenUri);
        $payload = $this->api->runReport($accessToken, $propertyId, $requestBody);

        if ($debug) {
            $debugWriter(
                'Debug payload for Google Analytics hourly runReport:',
                $this->api->encodeDebugJson($payload)
            );
        }

        $rows = [];
        foreach ((array) data_get($payload, 'rows', []) as $row) {
            $hourRaw = (string) data_get($row, 'dimensionValues.0.value', '');
            $timeslotLocal = $this->api->parseDateHourDimension($hourRaw, $timezone);
            if ($timeslotLocal === null || $timeslotLocal->lessThan($sinceLocal) || ! $timeslotLocal->lessThan($untilLocal)) {
                continue;
            }

            $rows[] = [
                'timeslot' => $timeslotLocal->utc()->format('Y-m-d H:i:s'),
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
            'db' => $this->persistRows($rows),
        ];
    }

    private function persistRows(array $rows): array
    {
        if (empty($rows)) {
            return ['inserted' => 0, 'updated' => 0, 'skipped' => 0];
        }

        $timeslots = array_values(array_unique(array_map(static fn (array $row): string => (string) $row['timeslot'], $rows)));
        $existing = StatHourlyGa::query()
            ->toBase()
            ->select(array_merge(['timeslot'], StatHourlyGa::COLUMN_NAMES))
            ->whereIn('timeslot', $timeslots)
            ->get();

        $existingByTimeslot = [];
        foreach ($existing as $item) {
            $existingByTimeslot[(string) $item->timeslot] = [
                'active_users' => (int) $item->active_users,
                'screen_page_views' => (int) $item->screen_page_views,
                'sessions' => (int) $item->sessions,
            ];
        }

        $upsertRows = [];
        $inserted = 0;
        $updated = 0;
        $skipped = 0;

        foreach ($rows as $row) {
            $timeslot = (string) $row['timeslot'];
            $current = [
                'active_users' => (int) $row['active_users'],
                'screen_page_views' => (int) $row['screen_page_views'],
                'sessions' => (int) $row['sessions'],
            ];
            $old = $existingByTimeslot[$timeslot] ?? null;

            if ($old === null) {
                $inserted++;
                $upsertRows[] = $row;

                continue;
            }

            if ($old !== $current) {
                $updated++;
                $upsertRows[] = $row;
            } else {
                $skipped++;
            }
        }

        if (! empty($upsertRows)) {
            StatHourlyGa::query()->upsert($upsertRows, ['timeslot'], StatHourlyGa::COLUMN_NAMES);
        }

        return ['inserted' => $inserted, 'updated' => $updated, 'skipped' => $skipped];
    }

    private function metricValue(array $row, int $index): int
    {
        $value = data_get($row, "metricValues.{$index}.value");

        return is_numeric($value) ? (int) $value : 0;
    }
}
