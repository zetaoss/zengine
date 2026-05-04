<?php

namespace App\Services;

use App\Jobs\DocTaskProcessJob;
use App\Models\DocTask;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Http;
use RuntimeException;
use Throwable;

class DocTaskService
{
    public const STATUS_PENDING = 'pending';

    public const STATUS_PROCESSING = 'processing';

    public const STATUS_RETRY_WAIT = 'retry_wait';

    public const STATUS_COMPLETED = 'completed';

    public const PHASE_RUNNING = 'running';

    public const PHASE_SLEEPING = 'sleeping';

    public const PHASE_RETRY_WAIT = 'retry_wait';

    private const STATE_CACHE_KEY = 'docfac:state';

    private const RETRY_AFTER_CACHE_KEY = 'docfac:retry_after';

    private const CREATE_PROMPT_TITLE = '틀:문서공장 생성 프롬프트';

    private const EDIT_PROMPT_TITLE = '틀:문서공장 편집 프롬프트';

    private const RETRY_BACKOFF = 1.1;

    private const PROCESSING_TIMEOUT_SECONDS = 300;

    private const PROCESS_LOCK_KEY = 'docfac:process-lock';

    private const PROCESS_LOCK_SECONDS = 120;

    private readonly int $intervalSeconds;
    private readonly int $retryIntervalSeconds;
    private readonly bool $autoEnqueue;
    private readonly string $mediaWikiApiServer;
    private readonly LLMService $llm;

    public function __construct(LLMService $llm)
    {
        $this->intervalSeconds = config('services.docfac.interval_seconds');
        $this->retryIntervalSeconds = config('services.docfac.retry_interval_seconds');
        $this->autoEnqueue = config('services.docfac.auto_enqueue');
        $this->mediaWikiApiServer = config('services.mediawiki.api_server');
        $this->llm = $llm;
    }

    public function processTask(): ?DocTask
    {
        $lock = Cache::store('redis')->lock(self::PROCESS_LOCK_KEY, self::PROCESS_LOCK_SECONDS);
        if (! $lock->get()) {
            return null;
        }

        try {
            return $this->processTaskCore();
        } finally {
            $lock->release();
        }
    }

    private function processTaskCore(): ?DocTask
    {
        if (! $this->isDueToRun()) {
            return null;
        }

        $task = $this->reserveHeadTask();
        if (! $task) {
            $created = $this->autoEnqueue ? $this->enqueueRecommendedWriteRequest() : null;
            if ($created) {
                $this->saveFactoryState([
                    'phase' => self::PHASE_SLEEPING,
                    'next_run_at' => null,
                    'task_id' => null,
                    'retry_count' => 0,
                    'last_error' => null,
                ]);

                $task = $this->reserveHeadTask();
            }
        }

        if (! $task) {
            $this->saveFactoryState([
                'phase' => self::PHASE_SLEEPING,
                'next_run_at' => now()->addSeconds($this->intervalSeconds)->toIso8601String(),
                'task_id' => null,
                'retry_count' => 0,
                'last_error' => null,
            ]);

            return null;
        }

        try {
            $content = $this->generateContent($task);
            $task->forceFill([
                'content' => $content,
                'status' => self::STATUS_COMPLETED,
                'skip_count' => 0,
                'last_error' => null,
            ])->save();

            $this->saveFactoryState([
                'phase' => self::PHASE_SLEEPING,
                'next_run_at' => now()->addSeconds($this->intervalSeconds)->toIso8601String(),
                'task_id' => null,
                'retry_count' => 0,
                'last_error' => null,
            ]);
        } catch (Throwable $e) {
            $message = mb_substr($e->getMessage(), 0, 2000, 'UTF-8');
            $retryCount = $this->nextRetryCount($task);

            $task->forceFill([
                'status' => self::STATUS_RETRY_WAIT,
                'error_count' => (int) $task->error_count + 1,
                'last_error' => $message,
            ])->save();

            $this->saveFactoryState([
                'phase' => self::PHASE_RETRY_WAIT,
                'next_run_at' => now()->addSeconds($this->retryDelaySeconds($retryCount))->toIso8601String(),
                'task_id' => $task->id,
                'retry_count' => $retryCount,
                'last_error' => $message,
            ]);

            throw $e;
        }

        return $task;
    }

    public function getStatus(): array
    {
        $head = $this->headTask();
        $state = $this->factoryState();

        return [
            ...$this->factoryConfig(),
            ...$state,
            'message' => $this->factoryMessage($state, $head),
            'next_run_after_seconds' => $this->secondsUntil($state['next_run_at']),
            'head' => $head ? $this->headTaskPayload($head) : null,
        ];
    }

    public function resumeProcessing(): array
    {
        Cache::store('redis')->forget(self::RETRY_AFTER_CACHE_KEY);

        $head = $this->headTask();
        if ($head && $head->status === self::STATUS_RETRY_WAIT) {
            $head->forceFill([
                'status' => self::STATUS_PENDING,
            ])->save();
        }

        $this->saveFactoryState([
            'phase' => self::PHASE_SLEEPING,
            'next_run_at' => null,
            'task_id' => null,
            'retry_count' => 0,
            'last_error' => null,
        ]);

        return $this->getStatus();
    }

    public function runNow(): array
    {
        Cache::store('redis')->forget(self::RETRY_AFTER_CACHE_KEY);
        $state = $this->factoryState();

        if ($state['phase'] === self::PHASE_RUNNING) {
            return $this->getStatus();
        }

        $nextRunAt = now()->startOfSecond()->toIso8601String();
        $this->saveFactoryState([
            'phase' => self::PHASE_SLEEPING,
            'next_run_at' => $nextRunAt,
            'task_id' => null,
            'retry_count' => 0,
            'last_error' => null,
        ]);

        DocTaskProcessJob::dispatch();

        return $this->getStatus();
    }

    private function reserveHeadTask(): ?DocTask
    {
        return DB::transaction(function () {
            $state = $this->factoryState();
            $task = $this->taskToRunQuery($state)
                ->lockForUpdate()
                ->first();

            if (! $task) {
                return null;
            }

            $now = now();

            if ($task->status === self::STATUS_PROCESSING && $task->updated_at && $task->updated_at->gt($now->copy()->subSeconds(self::PROCESSING_TIMEOUT_SECONDS))) {
                return null;
            }

            $task->forceFill([
                'status' => self::STATUS_PROCESSING,
                'attempts' => (int) $task->attempts + 1,
            ])->save();

            $this->saveFactoryState([
                'phase' => self::PHASE_RUNNING,
                'next_run_at' => null,
                'task_id' => $task->id,
                'retry_count' => $state['task_id'] === $task->id ? $state['retry_count'] : 0,
                'last_error' => $state['task_id'] === $task->id ? $state['last_error'] : null,
            ]);

            return $task;
        });
    }

    private function headTask(): ?DocTask
    {
        return $this->headTaskQuery()->first();
    }

    private function headTaskQuery()
    {
        return DocTask::query()
            ->where('status', '!=', self::STATUS_COMPLETED)
            ->orderBy('id');
    }

    private function headTaskPayload(DocTask $task): array
    {
        return [
            'id' => $task->id,
            'title' => $task->title,
            'status' => $task->status,
            'attempts' => $task->attempts,
            'error_count' => $task->error_count,
            'skip_count' => $task->skip_count,
            'last_error' => $task->last_error,
        ];
    }

    private function taskToRunQuery(array $state)
    {
        $query = DocTask::query()
            ->where('status', '!=', self::STATUS_COMPLETED);

        if ($state['phase'] === self::PHASE_RETRY_WAIT && $state['task_id']) {
            return $query->whereKey($state['task_id']);
        }

        return $query->orderBy('id');
    }

    private function isDueToRun(): bool
    {
        $state = $this->factoryState();

        if ($state['phase'] === self::PHASE_RUNNING) {
            $task = $state['task_id'] ? DocTask::query()->find($state['task_id']) : null;

            return $task?->updated_at
                ? $task->updated_at->lessThanOrEqualTo(now()->subSeconds(self::PROCESSING_TIMEOUT_SECONDS))
                : true;
        }

        if (! $state['next_run_at']) {
            return true;
        }

        return now()->greaterThanOrEqualTo(Carbon::parse($state['next_run_at']));
    }

    private function factoryConfig(): array
    {
        return [
            'interval' => $this->formatDuration($this->intervalSeconds),
            'retry_interval' => $this->formatDuration($this->retryIntervalSeconds),
            'retry_backoff' => self::RETRY_BACKOFF,
        ];
    }

    private function factoryState(): array
    {
        $state = Cache::store('redis')->get(self::STATE_CACHE_KEY);
        if (! is_array($state)) {
            $retryAfter = (int) (Cache::store('redis')->get(self::RETRY_AFTER_CACHE_KEY) ?? 0);

            return [
                'phase' => $retryAfter > now()->timestamp ? self::PHASE_RETRY_WAIT : self::PHASE_SLEEPING,
                'next_run_at' => $retryAfter > now()->timestamp ? date('c', $retryAfter) : null,
                'task_id' => $retryAfter > now()->timestamp ? $this->headTask()?->id : null,
                'retry_count' => 0,
                'last_error' => null,
            ];
        }

        return [
            'phase' => in_array($state['phase'] ?? null, [self::PHASE_RUNNING, self::PHASE_SLEEPING, self::PHASE_RETRY_WAIT], true)
                ? $state['phase']
                : self::PHASE_SLEEPING,
            'next_run_at' => $state['next_run_at'] ?? null,
            'task_id' => isset($state['task_id']) ? (int) $state['task_id'] : null,
            'retry_count' => isset($state['retry_count']) ? max(0, (int) $state['retry_count']) : 0,
            'last_error' => isset($state['last_error']) && is_string($state['last_error']) ? $state['last_error'] : null,
        ];
    }

    private function saveFactoryState(array $state): void
    {
        Cache::store('redis')->forever(self::STATE_CACHE_KEY, [
            'phase' => $state['phase'],
            'next_run_at' => $state['next_run_at'],
            'task_id' => $state['task_id'],
            'retry_count' => $state['retry_count'],
            'last_error' => $state['last_error'],
        ]);
    }

    private function factoryMessage(array $state, ?DocTask $head): string
    {
        if (! $head) {
            return '대기 중인 작업이 없습니다.';
        }

        if ($state['phase'] === self::PHASE_RUNNING) {
            return '작업을 처리 중입니다.';
        }

        if ($state['phase'] === self::PHASE_RETRY_WAIT) {
            return '오류 후 같은 작업의 재시도를 기다리는 중입니다.';
        }

        if ($state['next_run_at'] && now()->lessThan(Carbon::parse($state['next_run_at']))) {
            return '다음 작업 실행을 기다리는 중입니다.';
        }

        return '다음 실행 때 선두 작업을 처리합니다.';
    }

    private function secondsUntil(?string $nextRunAt): int
    {
        if (! $nextRunAt) {
            return 0;
        }

        return max(0, (int) now()->diffInSeconds(Carbon::parse($nextRunAt), false));
    }

    private function nextRetryCount(DocTask $task): int
    {
        $state = $this->factoryState();

        if ($state['task_id'] === $task->id) {
            return $state['retry_count'] + 1;
        }

        return 1;
    }

    private function retryDelaySeconds(int $retryCount): int
    {
        return (int) ceil($this->retryIntervalSeconds * (self::RETRY_BACKOFF ** max(0, $retryCount - 1)));
    }

    private function enqueueRecommendedWriteRequest(): ?DocTask
    {
        $row = DB::table('write_requests as w')
            ->leftJoin('doc_tasks as d', 'd.title', '=', 'w.title')
            ->select(['w.id', 'w.user_id', 'w.user_name', 'w.title'])
            ->selectRaw('(SELECT COALESCE(n.hit, 0) FROM not_matches n WHERE n.title = w.title LIMIT 1) as hit')
            ->whereNull('w.writed_at')
            ->whereNull('d.id')
            ->orderByRaw('w.rate DESC, hit DESC, w.ref DESC, w.id DESC')
            ->first();

        if (! $row) {
            return null;
        }

        return DocTask::create([
            'user_id' => (int) $row->user_id,
            'user_name' => (string) $row->user_name,
            'title' => (string) $row->title,
            'request_type' => 'create',
            'content' => '',
            'status' => self::STATUS_PENDING,
        ]);
    }

    private function generateContent(DocTask $task): string
    {
        $promptTitle = $task->request_type === 'edit'
            ? self::EDIT_PROMPT_TITLE
            : self::CREATE_PROMPT_TITLE;
        $promptTemplate = $this->fetchWikiRawText($promptTitle);
        $existingDoc = $task->request_type === 'edit'
            ? $this->fetchWikiRawText($task->title)
            : '';
        $prompt = str_replace(
            ['{제목}', '{기존문서}'],
            [$task->title, $existingDoc],
            $promptTemplate
        );

        return $this->llm->chat([
            ['role' => 'user', 'content' => $prompt],
        ]);
    }

    private function fetchWikiRawText(string $title): string
    {
        if ($this->mediaWikiApiServer === '') {
            throw new RuntimeException('Missing API_SERVER environment variable.');
        }

        $response = Http::acceptJson()
            ->timeout(20)
            ->get($this->mediaWikiApiServer.'/w/api.php', [
                'action' => 'query',
                'format' => 'json',
                'formatversion' => '2',
                'prop' => 'revisions',
                'rvprop' => 'content',
                'rvslots' => 'main',
                'titles' => $title,
            ]);

        if (! $response->ok()) {
            throw new RuntimeException("MediaWiki API request failed: HTTP {$response->status()} {$response->body()}");
        }

        $payload = $response->json();
        if (! is_array($payload)) {
            throw new RuntimeException('MediaWiki API returned invalid JSON.');
        }

        $page = data_get($payload, 'query.pages.0');
        if (! is_array($page) || array_key_exists('missing', $page)) {
            throw new RuntimeException("Prompt page not found: {$title}");
        }

        $content = data_get($page, 'revisions.0.slots.main.content') ?? data_get($page, 'revisions.0.content');
        if (! is_string($content) || trim($content) === '') {
            throw new RuntimeException("Prompt page is empty: {$title}");
        }

        return $content;
    }

    private function formatDuration(int $seconds): string
    {
        if ($seconds % 3600 === 0) {
            return ($seconds / 3600).'h';
        }

        if ($seconds % 60 === 0) {
            return ($seconds / 60).'m';
        }

        return $seconds.'s';
    }
}
