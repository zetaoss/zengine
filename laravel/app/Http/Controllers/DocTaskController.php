<?php

namespace App\Http\Controllers;

use App\Models\DocTask;
use App\Models\Page;
use App\Models\WriteRequest;
use App\Services\DocTaskService;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Http;

class DocTaskController extends Controller
{
    public function index()
    {
        $result = DocTask::query()
            ->select(['id', 'user_id', 'user_name', 'title', 'request_type', 'phase', 'created_at', 'updated_at'])
            ->orderByDesc('id')
            ->paginate(25);

        $result->getCollection()->transform(fn (DocTask $task) => $this->serializeTask($task));

        return $result;
    }

    public function status(DocTaskService $service)
    {
        return $service->getStatus();
    }

    public function resume(DocTaskService $service)
    {
        return $service->resumeProcessing();
    }

    public function runNow(DocTaskService $service)
    {
        return $service->runNow();
    }

    public function show(DocTask $docTask)
    {
        return $this->serializeTask($docTask);
    }

    public function clone(DocTask $docTask)
    {
        $clone = DocTask::create([
            'user_id' => auth()->id(),
            'user_name' => auth()->user()->name,
            'title' => $docTask->title,
            'request_type' => $docTask->request_type ?: 'create',
            'content' => $docTask->content,
            'phase' => DocTaskService::TASK_PHASE_PENDING,
            'attempts' => 0,
            'error_count' => 0,
            'skip_count' => 0,
            'last_error' => null,
        ]);

        return $this->serializeTask($clone);
    }

    public function storeFromWriteRequest(WriteRequest $writeRequest)
    {
        $existingTask = DocTask::query()
            ->where('title', $writeRequest->title)
            ->whereNotIn('phase', [DocTaskService::TASK_PHASE_SUCCEEDED, DocTaskService::TASK_PHASE_FAILED])
            ->orderByDesc('id')
            ->first();

        if ($existingTask) {
            return [
                'ok' => true,
                'id' => $existingTask->id,
                'created' => false,
            ];
        }

        $docTask = DocTask::create([
            'user_id' => auth()->id(),
            'user_name' => auth()->user()->name,
            'title' => $writeRequest->title,
            'request_type' => 'create',
            'content' => '',
            'phase' => DocTaskService::TASK_PHASE_PENDING,
        ]);

        return [
            'ok' => true,
            'id' => $docTask->id,
            'created' => true,
        ];
    }

    public function storeFromPage(Request $request)
    {
        $validated = $request->validate([
            'page_id' => 'required|integer|min:1',
            'request_type' => 'required|string|in:create,edit',
        ]);

        $title = $this->resolvePageTitle((int) $validated['page_id']);
        if (! is_string($title) || trim($title) === '') {
            return response()->json(['message' => '문서를 찾을 수 없습니다.'], 404);
        }

        $docTask = DocTask::create([
            'user_id' => auth()->id(),
            'user_name' => auth()->user()->name,
            'title' => $title,
            'request_type' => (string) $validated['request_type'],
            'content' => '',
            'phase' => DocTaskService::TASK_PHASE_PENDING,
        ]);

        return [
            'ok' => true,
            'id' => $docTask->id,
            'created' => true,
        ];
    }

    public function destroy(DocTask $docTask)
    {
        $docTask->delete();

        return ['ok' => true];
    }

    private function resolvePageTitle(int $pageId): ?string
    {
        $apiServer = (string) config('services.mediawiki.api_server');
        if ($apiServer !== '') {
            try {
                $response = Http::acceptJson()
                    ->timeout(10)
                    ->get($apiServer.'/w/api.php', [
                        'action' => 'query',
                        'format' => 'json',
                        'formatversion' => '2',
                        'pageids' => $pageId,
                    ]);

                if ($response->ok()) {
                    $title = data_get($response->json(), 'query.pages.0.title');
                    if (is_string($title) && trim($title) !== '') {
                        return str_replace('_', ' ', trim($title));
                    }
                }
            } catch (\Throwable) {
            }
        }

        $pageTitle = Page::query()
            ->whereKey($pageId)
            ->value('page_title');

        if (! is_string($pageTitle) || trim($pageTitle) === '') {
            return null;
        }

        return str_replace('_', ' ', trim($pageTitle));
    }

    private function serializeTask(DocTask $task): array
    {
        return [
            'id' => $task->id,
            'user_id' => $task->user_id,
            'user_name' => $task->user_name,
            'title' => $task->title,
            'request_type' => $task->request_type,
            'phase' => $task->phase,
            'content' => $task->content,
            'llm_model' => $task->llm_model,
            'attempts' => $task->attempts,
            'error_count' => $task->error_count,
            'skip_count' => $task->skip_count,
            'last_error' => $task->last_error,
            'created_at' => $task->created_at,
            'updated_at' => $task->updated_at,
        ];
    }
}
