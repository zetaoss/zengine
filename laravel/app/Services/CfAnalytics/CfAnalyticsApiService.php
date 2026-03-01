<?php

namespace App\Services\CfAnalytics;

use Carbon\CarbonImmutable;
use Illuminate\Support\Facades\Http;
use RuntimeException;

class CfAnalyticsApiService
{
    private const KST_TIMEZONE = 'Asia/Seoul';

    public function resolveCredentials(): array
    {
        $apiToken = (string) config('services.cloudflare.api_token');
        $zoneId = (string) config('services.cloudflare.zone_id');

        if ($apiToken === '' || $zoneId === '') {
            throw new RuntimeException('Missing Cloudflare credentials. Set CLOUDFLARE_API_TOKEN and CLOUDFLARE_ZONE_ID.');
        }

        return [$apiToken, $zoneId];
    }

    public function kstRangeForDays(int $days): array
    {
        if ($days < 1) {
            throw new RuntimeException('--days must be an integer greater than or equal to 1.');
        }

        $todayStartKst = CarbonImmutable::now(self::KST_TIMEZONE)->startOfDay();
        $sinceKst = $todayStartKst->subDays($days - 1);
        $untilKst = $todayStartKst->addDay();

        return [$sinceKst, $untilKst];
    }

    public function utcRangeForDays(int $days): array
    {
        if ($days < 1) {
            throw new RuntimeException('--days must be an integer greater than or equal to 1.');
        }

        $todayStartUtc = CarbonImmutable::now('UTC')->startOfDay();
        $sinceUtc = $todayStartUtc->subDays($days - 1);
        $untilUtc = $todayStartUtc->addDay();

        return [$sinceUtc, $untilUtc];
    }

    public function utcHourAnchoredRangeForDays(int $days): array
    {
        if ($days < 1) {
            throw new RuntimeException('--days must be an integer greater than or equal to 1.');
        }

        $untilUtc = CarbonImmutable::now('UTC')->startOfHour();
        $sinceUtc = $untilUtc->subDays($days);

        return [$sinceUtc, $untilUtc];
    }

    public function kstDateStrings(CarbonImmutable $sinceKst, CarbonImmutable $untilKst): array
    {
        $dates = [];
        $cursor = $sinceKst;
        while ($cursor->lessThan($untilKst)) {
            $dates[] = $cursor->toDateString();
            $cursor = $cursor->addDay();
        }

        return $dates;
    }

    public function kstDayWindows(CarbonImmutable $sinceKst, CarbonImmutable $untilKst): array
    {
        $windows = [];
        $cursor = $sinceKst;
        while ($cursor->lessThan($untilKst)) {
            $startKst = $cursor;
            $endKst = $cursor->addDay();
            $windows[] = [
                'date_kst' => $startKst->toDateString(),
                'start_utc' => $startKst->utc(),
                'end_utc' => $endKst->utc(),
            ];
            $cursor = $endKst;
        }

        return $windows;
    }

    public function utcDayWindows(CarbonImmutable $sinceUtc, CarbonImmutable $untilUtc): array
    {
        $windows = [];
        $cursor = $sinceUtc;
        while ($cursor->lessThan($untilUtc)) {
            $startUtc = $cursor;
            $endUtc = $cursor->addDay();
            $windows[] = [
                'date_utc' => $startUtc->toDateString(),
                'start_utc' => $startUtc->utc(),
                'end_utc' => $endUtc->utc(),
            ];
            $cursor = $endUtc;
        }

        return $windows;
    }

    public function runGraphql(string $apiToken, string $query, array $variables): array
    {
        $response = Http::withToken($apiToken)
            ->acceptJson()
            ->asJson()
            ->timeout(20)
            ->post('https://api.cloudflare.com/client/v4/graphql', [
                'query' => $query,
                'variables' => $variables,
            ]);

        if (! $response->ok()) {
            throw new RuntimeException("Cloudflare API request failed: HTTP {$response->status()} {$response->body()}");
        }

        $payload = $response->json();
        $errors = data_get($payload, 'errors', []);
        if (! empty($errors)) {
            $messages = [];
            foreach ($errors as $error) {
                $messages[] = (string) ($error['message'] ?? 'Unknown error');
            }
            throw new RuntimeException('Cloudflare GraphQL returned errors: '.implode(' | ', $messages));
        }

        return $payload;
    }

    public function encodeDebugJson(array $payload): string
    {
        $json = json_encode($payload, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);

        return $json === false ? '{}' : $json;
    }

    public function todayKstDateString(): string
    {
        return CarbonImmutable::now(self::KST_TIMEZONE)->toDateString();
    }
}
