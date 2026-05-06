<?php

namespace App\Services;

use App\Jobs\DocTaskProcessJob;
use App\Models\DocTask;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Http;
use RuntimeException;
use Throwable;

class DocTaskService
{
    public const TASK_PHASE_PENDING = 'Pending';

    public const TASK_PHASE_RUNNING = 'Running';

    public const TASK_PHASE_BACKOFF = 'Backoff';

    public const TASK_PHASE_FAILED = 'Failed';

    public const TASK_PHASE_SUCCEEDED = 'Succeeded';

    public const STATUS_RUNNING = 'Running';

    public const STATUS_WAITING = 'Waiting';

    public const STATUS_BACKOFF = 'Backoff';

    private const PHASE_CACHE_KEY = 'docfac:phase';

    private const CREATE_PROMPT_TEMPLATE_TITLE = '틀:문서공장 생성 프롬프트';

    private const EDIT_PROMPT_TEMPLATE_TITLE = '틀:문서공장 편집 프롬프트';

    private const RETRY_BACKOFF_MULTIPLIER = 1.1;

    private const TASK_RUNNING_TIMEOUT_SECONDS = 300;

    private const PROCESS_LOCK_KEY = 'docfac:process-lock';

    private const PROCESS_LOCK_TTL_SECONDS = 120;

    private readonly int $autoEnqueueIntervalSeconds;
    private readonly int $processIntervalSeconds;
    private readonly int $retryIntervalSeconds;
    private readonly int $maxRetries;
    private readonly string $mediaWikiApiServer;
    private readonly LLMService $llm;

    public function __construct(LLMService $llm)
    {
        $this->autoEnqueueIntervalSeconds = config('services.docfac.autoenqueue_interval');
        $this->processIntervalSeconds = config('services.docfac.process_interval');
        $this->retryIntervalSeconds = config('services.docfac.retry_interval');
        $this->maxRetries = config('services.docfac.max_retries');
        $this->mediaWikiApiServer = config('services.mediawiki.api_server');
        $this->llm = $llm;
    }

    public function processTask(): ?DocTask
    {
        $lock = Cache::store('redis')->lock(self::PROCESS_LOCK_KEY, self::PROCESS_LOCK_TTL_SECONDS);
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
        $state = $this->factoryState();
        $head = $this->headTask();

        // One scheduler call is one tick. Accumulate ticks while waiting/backoff.
        if (in_array($state['status'], [self::STATUS_WAITING, self::STATUS_BACKOFF], true)) {
            $state['tick_count'] += 1;
            $this->saveFactoryState($state);
        }

        if (! $this->isDueToRun($state, $head)) {
            return null;
        }

        $task = $this->reserveHeadTask();
        if (! $task) {
            $created = $this->shouldAutoEnqueue() ? $this->enqueueRecommendedWriteRequest() : null;
            if ($created) {
                $this->saveFactoryState([
                    'status' => self::STATUS_WAITING,
                    'task_id' => null,
                    'retry_count' => 0,
                    'last_error' => null,
                    'tick_count' => 0,
                ]);

                $task = $this->reserveHeadTask();
            }
        }

        if (! $task) {
            $this->saveFactoryState([
                'status' => self::STATUS_WAITING,
                'task_id' => null,
                'retry_count' => 0,
                'last_error' => null,
                'tick_count' => 0,
            ]);

            return null;
        }

        try {
            $result = $this->generateContent($task);
            $task->forceFill([
                'content' => $result['content'],
                'llm_model' => $result['model'],
                'phase' => self::TASK_PHASE_SUCCEEDED,
                'skip_count' => 0,
                'last_error' => null,
            ])->save();

            $this->saveFactoryState([
                'status' => self::STATUS_WAITING,
                'task_id' => null,
                'retry_count' => 0,
                'last_error' => null,
                'tick_count' => 0,
            ]);
        } catch (Throwable $e) {
            $message = mb_substr($e->getMessage(), 0, 2000, 'UTF-8');
            $retryCount = $this->nextRetryCount($task);

            if ($retryCount >= $this->maxRetries) {
                $task->forceFill([
                    'phase' => self::TASK_PHASE_FAILED,
                    'error_count' => (int) $task->error_count + 1,
                    'last_error' => $message,
                ])->save();

                $this->saveFactoryState([
                    'status' => self::STATUS_WAITING,
                    'task_id' => null,
                    'retry_count' => 0,
                    'last_error' => null,
                    'tick_count' => 0,
                ]);

                return $task;
            }

            $task->forceFill([
                'phase' => self::TASK_PHASE_BACKOFF,
                'error_count' => (int) $task->error_count + 1,
                'last_error' => $message,
            ])->save();

            $this->saveFactoryState([
                'status' => self::STATUS_BACKOFF,
                'task_id' => $task->id,
                'retry_count' => $retryCount,
                'last_error' => $message,
                'tick_count' => 0,
            ]);

            throw $e;
        }

        return $task;
    }

    public function getStatus(): array
    {
        $head = $this->headTask();
        $state = $this->factoryState();
        $nextRunAfterSeconds = $this->nextRunAfterSeconds($state, $head);

        return [
            ...$this->factoryConfig(),
            'status' => $state['status'],
            'message' => $this->factoryMessage($state, $head),
            'next_run_after_seconds' => $nextRunAfterSeconds,
            'next_run_at' => $this->nextRunAt($state, $nextRunAfterSeconds),
            'task_id' => $state['task_id'],
            'last_error' => $state['last_error'],
            'head' => $head ? $this->headTaskPayload($head) : null,
        ];
    }

    public function resumeProcessing(): array
    {
        $head = $this->headTask();
        if ($head && $head->phase === self::TASK_PHASE_BACKOFF) {
            $head->forceFill([
                'phase' => self::TASK_PHASE_PENDING,
            ])->save();
        }

        $this->saveFactoryState([
            'status' => self::STATUS_WAITING,
            'task_id' => null,
            'retry_count' => 0,
            'last_error' => null,
            'tick_count' => 0,
        ]);

        return $this->getStatus();
    }

    public function runNow(): array
    {
        $state = $this->factoryState();

        if ($state['status'] === self::STATUS_RUNNING) {
            return $this->getStatus();
        }

        if ($state['status'] === self::STATUS_BACKOFF && $state['task_id']) {
            // Keep retry chain context, but make it due immediately.
            $this->saveFactoryState([
                'status' => self::STATUS_BACKOFF,
                'task_id' => $state['task_id'],
                'retry_count' => $state['retry_count'],
                'last_error' => $state['last_error'],
                'tick_count' => $this->retryIntervalTicks(max(1, $state['retry_count'])),
            ]);
        } else {
            $this->saveFactoryState([
                'status' => self::STATUS_WAITING,
                'task_id' => null,
                'retry_count' => 0,
                'last_error' => null,
                'tick_count' => $this->processIntervalTicks(),
            ]);
        }

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

            if ($task->phase === self::TASK_PHASE_RUNNING && $task->updated_at && $task->updated_at->gt($now->copy()->subSeconds(self::TASK_RUNNING_TIMEOUT_SECONDS))) {
                return null;
            }

            $task->forceFill([
                'phase' => self::TASK_PHASE_RUNNING,
                'attempts' => (int) $task->attempts + 1,
            ])->save();

            $this->saveFactoryState([
                'status' => self::STATUS_RUNNING,
                'task_id' => $task->id,
                'retry_count' => $state['task_id'] === $task->id ? $state['retry_count'] : 0,
                'last_error' => $state['task_id'] === $task->id ? $state['last_error'] : null,
                'tick_count' => 0,
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
            ->whereNotIn('phase', [self::TASK_PHASE_SUCCEEDED, self::TASK_PHASE_FAILED])
            ->orderBy('id');
    }

    private function headTaskPayload(DocTask $task): array
    {
        return [
            'id' => $task->id,
            'title' => $task->title,
            'phase' => $task->phase,
            'attempts' => $task->attempts,
            'error_count' => $task->error_count,
            'skip_count' => $task->skip_count,
            'last_error' => $task->last_error,
        ];
    }

    private function taskToRunQuery(array $state)
    {
        $query = DocTask::query()
            ->whereNotIn('phase', [self::TASK_PHASE_SUCCEEDED, self::TASK_PHASE_FAILED]);

        if ($state['status'] === self::STATUS_BACKOFF && $state['task_id']) {
            return $query->whereKey($state['task_id']);
        }

        return $query->orderBy('id');
    }

    private function isDueToRun(array $state, ?DocTask $head): bool
    {
        if ($state['status'] === self::STATUS_RUNNING) {
            $task = $state['task_id'] ? DocTask::query()->find($state['task_id']) : null;

            return $task?->updated_at
                ? $task->updated_at->lessThanOrEqualTo(now()->subSeconds(self::TASK_RUNNING_TIMEOUT_SECONDS))
                : true;
        }

        if ($state['status'] === self::STATUS_BACKOFF) {
            return $state['tick_count'] >= $this->retryIntervalTicks($state['retry_count']);
        }

        // Waiting state
        if ($head) {
            return $state['tick_count'] >= $this->processIntervalTicks();
        }

        if ($this->shouldAutoEnqueue()) {
            return $state['tick_count'] >= $this->autoenqueueIntervalTicks();
        }

        return false;
    }

    private function factoryConfig(): array
    {
        return [
            'process_interval' => $this->formatDuration($this->processIntervalSeconds),
            'interval' => $this->formatDuration($this->processIntervalSeconds),
            'autoenqueue_interval' => $this->formatOptionalDuration($this->autoEnqueueIntervalSeconds),
            'retry_interval' => $this->formatDuration($this->retryIntervalSeconds),
            'retry_backoff' => self::RETRY_BACKOFF_MULTIPLIER,
        ];
    }

    private function factoryState(): array
    {
        $state = Cache::store('redis')->get(self::PHASE_CACHE_KEY);
        if (! is_array($state)) {
            return [
                'status' => self::STATUS_WAITING,
                'task_id' => null,
                'retry_count' => 0,
                'last_error' => null,
                'tick_count' => 0,
                'state_changed_at' => now()->toIso8601String(),
            ];
        }

        return [
            'status' => $this->normalizeStatus($state['status'] ?? null),
            'task_id' => isset($state['task_id']) ? (int) $state['task_id'] : null,
            'retry_count' => isset($state['retry_count']) ? max(0, (int) $state['retry_count']) : 0,
            'last_error' => isset($state['last_error']) && is_string($state['last_error']) ? $state['last_error'] : null,
            'tick_count' => isset($state['tick_count']) ? max(0, (int) $state['tick_count']) : 0,
            'state_changed_at' => isset($state['state_changed_at']) && is_string($state['state_changed_at'])
                ? $state['state_changed_at']
                : now()->toIso8601String(),
        ];
    }

    private function saveFactoryState(array $state): void
    {
        Cache::store('redis')->forever(self::PHASE_CACHE_KEY, [
            'status' => $state['status'],
            'task_id' => $state['task_id'],
            'retry_count' => $state['retry_count'],
            'last_error' => $state['last_error'],
            'tick_count' => $state['tick_count'] ?? 0,
            'state_changed_at' => $state['state_changed_at'] ?? now()->toIso8601String(),
        ]);
    }

    private function factoryMessage(array $state, ?DocTask $head): string
    {
        if (! $head) {
            return '대기 중인 작업이 없습니다.';
        }

        if ($state['status'] === self::STATUS_RUNNING) {
            return '작업을 처리 중입니다.';
        }

        if ($state['status'] === self::STATUS_BACKOFF) {
            return '오류 후 재시도를 기다리는 중입니다.';
        }

        if ($this->nextRunAfterSeconds($state, $head) > 0) {
            return '다음 작업 실행을 기다리는 중입니다.';
        }

        return '다음 실행 때 작업을 처리합니다.';
    }

    private function nextRetryCount(DocTask $task): int
    {
        $state = $this->factoryState();

        if ($state['task_id'] === $task->id) {
            return $state['retry_count'] + 1;
        }

        // Fallback to persisted task error history when runtime state was reset manually.
        if ($task->phase === self::TASK_PHASE_BACKOFF && (int) $task->error_count > 0) {
            return (int) $task->error_count + 1;
        }

        return 1;
    }

    private function retryDelaySeconds(int $retryCount): int
    {
        return (int) ceil($this->retryIntervalSeconds * (self::RETRY_BACKOFF_MULTIPLIER ** max(0, $retryCount - 1)));
    }

    private function shouldAutoEnqueue(): bool
    {
        return $this->autoEnqueueIntervalSeconds > 0;
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
            'phase' => self::TASK_PHASE_PENDING,
        ]);
    }

    private function generateContent(DocTask $task): array
    {
        $promptTitle = $task->request_type === 'edit'
            ? self::EDIT_PROMPT_TEMPLATE_TITLE
            : self::CREATE_PROMPT_TEMPLATE_TITLE;
        $promptTemplate = $this->fetchWikiRawText($promptTitle);
        $existingDoc = $task->request_type === 'edit'
            ? $this->fetchWikiRawText($task->title)
            : '';
        $prompt = str_replace(
            ['{제목}', '{기존문서}'],
            [$task->title, $existingDoc],
            $promptTemplate
        );

        return $this->llm->chatCompletion([
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

    private function formatOptionalDuration(int $seconds): string
    {
        if ($seconds <= 0) {
            return 'off';
        }

        return $this->formatDuration($seconds);
    }

    private function normalizeStatus(mixed $status): string
    {
        return match ($status) {
            self::STATUS_RUNNING => self::STATUS_RUNNING,
            self::STATUS_BACKOFF => self::STATUS_BACKOFF,
            self::STATUS_WAITING => self::STATUS_WAITING,
            default => self::STATUS_WAITING,
        };
    }

    private function processIntervalTicks(): int
    {
        return max(1, (int) ceil($this->processIntervalSeconds / 60));
    }

    private function autoenqueueIntervalTicks(): int
    {
        if ($this->autoEnqueueIntervalSeconds <= 0) {
            return 0;
        }

        return max(1, (int) ceil($this->autoEnqueueIntervalSeconds / 60));
    }

    private function retryIntervalTicks(int $retryCount): int
    {
        return max(1, (int) ceil($this->retryDelaySeconds(max(1, $retryCount)) / 60));
    }

    private function nextRunAfterSeconds(array $state, ?DocTask $head): int
    {
        if ($state['status'] === self::STATUS_RUNNING) {
            return 0;
        }

        if ($state['status'] === self::STATUS_BACKOFF) {
            $remainingTicks = max(0, $this->retryIntervalTicks($state['retry_count']) - $state['tick_count']);

            return $remainingTicks * 60;
        }

        // Waiting
        if ($head) {
            $remainingTicks = max(0, $this->processIntervalTicks() - $state['tick_count']);

            return $remainingTicks * 60;
        }

        if ($this->shouldAutoEnqueue()) {
            $remainingTicks = max(0, $this->autoenqueueIntervalTicks() - $state['tick_count']);

            return $remainingTicks * 60;
        }

        return 0;
    }

    private function nextRunAt(array $state, int $nextRunAfterSeconds): ?string
    {
        if ($nextRunAfterSeconds <= 0) {
            return null;
        }

        return now()->addSeconds($nextRunAfterSeconds)->toIso8601String();
    }
}
