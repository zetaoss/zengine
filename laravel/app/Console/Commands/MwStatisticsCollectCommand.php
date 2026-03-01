<?php

namespace App\Console\Commands;

use App\Models\MwStatistics;
use Carbon\CarbonImmutable;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Http;
use RuntimeException;

class MwStatisticsCollectCommand extends Command
{
    protected $signature = 'app:mw-statistics-collect
                            {--date= : Target KST date (YYYY-MM-DD), default is today KST}
                            {--debug : Print raw JSON payload}';
    protected $description = 'Collect MediaWiki site statistics and upsert daily snapshot';

    public function handle(): int
    {
        try {
            $timeslot = $this->resolveTargetDate();
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

            MwStatistics::query()->upsert([$row], ['timeslot'], [
                'pages',
                'articles',
                'edits',
                'images',
                'users',
                'activeusers',
                'admins',
                'jobs',
            ]);

            if ((bool) $this->option('debug')) {
                $this->line(json_encode($payload, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES) ?: '{}');
            }

            $this->info("Stored mw_statistics for timeslot={$timeslot}");
            $this->table(array_keys($row), [$row]);

            return Command::SUCCESS;
        } catch (RuntimeException $e) {
            $this->error($e->getMessage());

            return Command::FAILURE;
        }
    }

    private function resolveTargetDate(): string
    {
        $dateInput = trim((string) $this->option('date'));
        if ($dateInput === '') {
            return CarbonImmutable::now('Asia/Seoul')->toDateString();
        }

        try {
            return CarbonImmutable::parse($dateInput, 'Asia/Seoul')->toDateString();
        } catch (\Throwable) {
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
