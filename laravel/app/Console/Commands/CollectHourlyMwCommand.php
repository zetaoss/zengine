<?php

namespace App\Console\Commands;

use App\Models\StatHourlyMw;
use Carbon\CarbonImmutable;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Http;
use RuntimeException;

class CollectHourlyMwCommand extends Command
{
    protected $signature = 'z:collect-hourly-mw
                            {--at= : Target UTC hour (YYYY-MM-DDTHH:00:00Z), default is current UTC hour}
                            {--debug : Print raw JSON payload}';
    protected $description = 'Collect MediaWiki site statistics and upsert hourly snapshot';

    public function handle(): int
    {
        try {
            $timeslot = $this->resolveTargetHour();
            $payload = $this->fetchPayload();
            $statistics = (array) data_get($payload, 'query.statistics', []);

            if (empty($statistics)) {
                throw new RuntimeException('MediaWiki response is missing query.statistics.');
            }

            $row = [
                'timeslot' => $timeslot->format('Y-m-d H:i:s'),
                'pages' => (int) ($statistics['pages'] ?? 0),
                'articles' => (int) ($statistics['articles'] ?? 0),
                'edits' => (int) ($statistics['edits'] ?? 0),
                'images' => (int) ($statistics['images'] ?? 0),
                'users' => (int) ($statistics['users'] ?? 0),
                'activeusers' => (int) ($statistics['activeusers'] ?? 0),
                'admins' => (int) ($statistics['admins'] ?? 0),
                'jobs' => (int) ($statistics['jobs'] ?? 0),
            ];

            StatHourlyMw::query()->upsert([$row], ['timeslot'], [
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

            $this->info("Stored stat_hourly_mw row for timeslot={$timeslot->format('Y-m-d\\TH:i:s\\Z')}");
            $this->table(array_keys($row), [$row]);

            return Command::SUCCESS;
        } catch (RuntimeException $e) {
            $this->error($e->getMessage());

            return Command::FAILURE;
        }
    }

    private function resolveTargetHour(): CarbonImmutable
    {
        $atInput = trim((string) $this->option('at'));
        if ($atInput === '') {
            return CarbonImmutable::now('UTC')->startOfHour();
        }

        try {
            return CarbonImmutable::parse($atInput, 'UTC')->startOfHour()->utc();
        } catch (\Carbon\Exceptions\InvalidFormatException) {
            throw new RuntimeException('--at must be a valid UTC datetime such as 2026-03-14T15:00:00Z.');
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
