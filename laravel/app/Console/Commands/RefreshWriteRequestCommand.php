<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;
use Illuminate\Support\Collection;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Http;
use RuntimeException;

class RefreshWriteRequestCommand extends Command
{
    protected $signature = 'z:refresh-write-request {--dry-run : Print updates without writing to DB}';
    protected $description = 'Refresh write-request status for the first page of todo, todo-top, and done lists';

    public function handle(): int
    {
        try {
            $targets = $this->collectTargets();
            $targetIds = $targets->pluck('id')->map(fn ($id) => (int) $id)->all();
            if ($targets->isEmpty()) {
                $this->info('No write-request rows found in first pages.');

                return Command::SUCCESS;
            }

            $titleExists = $this->fetchTitleExistsMap($targets->pluck('title')->all());

            $now = now();
            $toDone = [];
            $toTodo = [];

            foreach ($targets as $row) {
                $title = (string) $row->title;
                $exists = $this->resolveExists($titleExists, $title);
                $isDone = ! is_null($row->writed_at);

                if ($exists && ! $isDone) {
                    $toDone[] = (int) $row->id;

                    continue;
                }

                if (! $exists && $isDone) {
                    $toTodo[] = (int) $row->id;
                }
            }

            if ((bool) $this->option('dry-run')) {
                $this->line('Dry-run mode enabled. No DB updates applied.');
            } else {
                if (! empty($toDone)) {
                    DB::table('write_requests')
                        ->whereIn('id', $toDone)
                        ->update([
                            'writer_id' => 0,
                            'writer_name' => 'Unknown',
                            'writed_at' => $now,
                            'updated_at' => $now,
                        ]);
                }

                if (! empty($toTodo)) {
                    DB::table('write_requests')
                        ->whereIn('id', $toTodo)
                        ->update([
                            'writed_at' => null,
                            'updated_at' => $now,
                        ]);
                }

                // Mark all checked rows as refreshed so stale sweep keeps rotating.
                if (! empty($targetIds)) {
                    DB::table('write_requests')
                        ->whereIn('id', $targetIds)
                        ->update([
                            'updated_at' => $now,
                        ]);
                }
            }

            $this->info('Write-request status refresh complete.');
            $this->table(
                ['checked', 'to_done', 'to_todo', 'dry_run'],
                [[(string) $targets->count(), (string) count($toDone), (string) count($toTodo), (bool) $this->option('dry-run') ? 'yes' : 'no']]
            );

            return Command::SUCCESS;
        } catch (RuntimeException $e) {
            $this->error($e->getMessage());

            return Command::FAILURE;
        }
    }

    private function collectTargets(): Collection
    {
        $todo = DB::table('write_requests')
            ->select(['id', 'title', 'writed_at'])
            ->whereNull('writed_at')
            ->orderByDesc('id')
            ->limit(50)
            ->get();

        $todoTop = DB::table('write_requests as w')
            ->select(['w.id', 'w.title', 'w.writed_at'])
            ->selectRaw('(SELECT COALESCE(n.hit, 0) FROM not_matches n WHERE n.title = w.title LIMIT 1) as hit')
            ->whereNull('w.writed_at')
            ->orderByRaw('w.rate DESC, hit DESC, w.ref DESC, w.id DESC')
            ->limit(50)
            ->get();

        $done = DB::table('write_requests')
            ->select(['id', 'title', 'writed_at'])
            ->whereNotNull('writed_at')
            ->orderByDesc('writed_at')
            ->orderByDesc('id')
            ->limit(50)
            ->get();

        $stale = DB::table('write_requests')
            ->select(['id', 'title', 'writed_at'])
            ->orderBy('updated_at')
            ->orderBy('id')
            ->limit(50)
            ->get();

        return $todo
            ->concat($todoTop)
            ->concat($done)
            ->concat($stale)
            ->unique('id')
            ->values();
    }

    private function fetchTitleExistsMap(array $titles): array
    {
        $apiServer = (string) config('services.mediawiki.api_server');

        $map = [];

        foreach (array_chunk(array_values(array_unique($titles)), 50) as $chunk) {
            $response = Http::acceptJson()
                ->timeout(20)
                ->get($apiServer.'/w/api.php', [
                    'action' => 'query',
                    'format' => 'json',
                    'titles' => implode('|', $chunk),
                ]);

            if (! $response->ok()) {
                throw new RuntimeException("MediaWiki API request failed: HTTP {$response->status()} {$response->body()}");
            }

            $payload = $response->json();
            if (! is_array($payload)) {
                throw new RuntimeException('MediaWiki API returned invalid JSON.');
            }

            $pages = data_get($payload, 'query.pages');
            if (! is_array($pages)) {
                throw new RuntimeException('MediaWiki API response is missing query.pages.');
            }

            foreach ($pages as $page) {
                if (! is_array($page)) {
                    continue;
                }
                $title = (string) ($page['title'] ?? '');
                if ($title === '') {
                    continue;
                }
                $exists = ! array_key_exists('missing', $page);
                foreach ($this->titleVariants($title) as $variant) {
                    $map[$this->normalizeTitleKey($variant)] = $exists;
                }
            }
        }

        return $map;
    }

    private function resolveExists(array $map, string $title): bool
    {
        foreach ($this->titleVariants($title) as $variant) {
            $normalized = $this->normalizeTitleKey($variant);
            if (array_key_exists($normalized, $map)) {
                return (bool) $map[$normalized];
            }
        }

        return false;
    }

    private function titleVariants(string $title): array
    {
        $title = trim($title);
        if ($title === '') {
            return [];
        }

        return array_values(array_unique([
            $title,
            str_replace('_', ' ', $title),
            str_replace(' ', '_', $title),
        ]));
    }

    private function normalizeTitleKey(string $title): string
    {
        return mb_strtolower(trim($title), 'UTF-8');
    }
}
