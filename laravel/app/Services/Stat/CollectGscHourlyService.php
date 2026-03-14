<?php

namespace App\Services\Stat;

use App\Models\StatGscHourly;

class CollectGscHourlyService
{
    public function __construct(
        private readonly CollectGscApiService $api,
        private readonly CollectGscPersistService $persist,
    ) {}

    public function collect(int $days, bool $debug, callable $debugWriter): array
    {
        [$siteUrl, $clientEmail, $privateKey, $timezone, $tokenUri] = $this->api->resolveCredentials();
        [$sinceLocal, $untilLocal] = $this->api->propertyHourAnchoredRangeForDays($days);

        $requestBody = [
            'startDate' => $sinceLocal->toDateString(),
            'endDate' => $untilLocal->subHour()->toDateString(),
            'dimensions' => ['HOUR'],
            'dataState' => 'HOURLY_ALL',
            'rowLimit' => 25000,
        ];

        if ($debug) {
            $debugWriter(
                'Debug request for Google Search Console hourly query:',
                $this->api->encodeDebugJson($requestBody)
            );
        }

        $accessToken = $this->api->fetchAccessToken($clientEmail, $privateKey, $tokenUri);
        $payload = $this->api->query($accessToken, $siteUrl, $requestBody);

        if ($debug) {
            $debugWriter(
                'Debug payload for Google Search Console hourly query:',
                $this->api->encodeDebugJson($payload)
            );
        }

        $rows = [];
        foreach ((array) data_get($payload, 'rows', []) as $row) {
            $dateHourRaw = (string) data_get($row, 'keys.0', '');
            $timeslotLocal = $this->api->parseHourDimension($dateHourRaw);
            if ($timeslotLocal === null || $timeslotLocal->lessThan($sinceLocal) || ! $timeslotLocal->lessThan($untilLocal)) {
                continue;
            }

            $rows[] = [
                'timeslot' => $timeslotLocal->utc()->format('Y-m-d H:i:s'),
                'clicks' => $this->intMetric($row, 'clicks'),
                'impressions' => $this->intMetric($row, 'impressions'),
                'ctr' => $this->percentMetric($row, 'ctr'),
                'position' => $this->floatMetric($row, 'position'),
            ];
        }

        return [
            'siteUrl' => $siteUrl,
            'timezone' => $timezone,
            'sinceLocal' => $sinceLocal,
            'untilLocal' => $untilLocal,
            'timeslots' => array_values(array_map(static fn (array $row): string => (string) $row['timeslot'], $rows)),
            'db' => $this->persist->persistRows(StatGscHourly::class, $rows),
        ];
    }

    private function intMetric(array $row, string $key): int
    {
        $value = data_get($row, $key);

        return is_numeric($value) ? (int) round((float) $value) : 0;
    }

    private function percentMetric(array $row, string $key): float
    {
        $value = data_get($row, $key);

        return is_numeric($value) ? round((float) $value * 100, 4) : 0.0;
    }

    private function floatMetric(array $row, string $key): float
    {
        $value = data_get($row, $key);

        return is_numeric($value) ? round((float) $value, 4) : 0.0;
    }
}
