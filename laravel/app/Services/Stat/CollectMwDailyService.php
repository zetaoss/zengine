<?php

namespace App\Services\Stat;

use App\Models\StatMwDaily;
use Carbon\CarbonImmutable;
use Illuminate\Support\Facades\Http;
use RuntimeException;

class CollectMwDailyService
{
    public function collect(string $dateInput = ''): array
    {
        $timeslot = $this->resolveTargetDate($dateInput);
        $payload = $this->fetchPayload();
        $statistics = (array) data_get($payload, 'query.statistics', []);

        if (empty($statistics)) {
            throw new RuntimeException('MediaWiki response is missing query.statistics.');
        }

        $row = [
            'timeslot' => $timeslot,
            'pages' => (int) ($statistics['pages'] ?? 0),
            'articles' => (int) ($statistics['articles'] ?? 0),
            'edits' => (int) ($statistics['edits'] ?? 0),
            'images' => (int) ($statistics['images'] ?? 0),
            'users' => (int) ($statistics['users'] ?? 0),
            'activeusers' => (int) ($statistics['activeusers'] ?? 0),
            'admins' => (int) ($statistics['admins'] ?? 0),
            'jobs' => (int) ($statistics['jobs'] ?? 0),
        ];

        StatMwDaily::query()->upsert([$row], ['timeslot'], [
            'pages',
            'articles',
            'edits',
            'images',
            'users',
            'activeusers',
            'admins',
            'jobs',
        ]);

        return [
            'payload' => $payload,
            'row' => $row,
        ];
    }

    private function resolveTargetDate(string $dateInput): string
    {
        if ($dateInput === '') {
            return CarbonImmutable::now('Asia/Seoul')->toDateString();
        }

        try {
            return CarbonImmutable::parse($dateInput, 'Asia/Seoul')->toDateString();
        } catch (\Carbon\Exceptions\InvalidFormatException) {
            throw new RuntimeException('--date must be in YYYY-MM-DD format.');
        }
    }

    private function fetchPayload(): array
    {
        $apiServer = rtrim((string) env('API_SERVER', ''), '/');
        if ($apiServer === '') {
            throw new RuntimeException('Missing API_SERVER environment variable.');
        }

        $response = Http::acceptJson()
            ->timeout(20)
            ->get($apiServer.'/w/api.php', [
                'action' => 'query',
                'meta' => 'siteinfo',
                'siprop' => 'statistics',
                'format' => 'json',
            ]);

        if (! $response->ok()) {
            throw new RuntimeException("MediaWiki API request failed: HTTP {$response->status()} {$response->body()}");
        }

        $payload = $response->json();
        if (! is_array($payload)) {
            throw new RuntimeException('MediaWiki API returned invalid JSON.');
        }

        return $payload;
    }
}
