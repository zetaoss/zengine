<?php

namespace App\Http\Controllers;

use App\Models\DocTask;
use App\Models\WriteRequest;
use App\Services\DocTaskService;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class DocTaskController extends Controller
{
    public function index()
    {
        return DocTask::query()
            ->select(['id', 'user_id', 'user_name', 'title', 'request_type', 'status', 'created_at', 'updated_at'])
            ->orderByDesc('id')
            ->paginate(25);
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
        return $docTask;
    }

    public function clone(DocTask $docTask)
    {
        $clone = DocTask::create([
            'user_id' => auth()->id(),
            'user_name' => auth()->user()->name,
            'title' => $docTask->title,
            'request_type' => $docTask->request_type ?: 'create',
            'content' => $docTask->content,
            'status' => DocTaskService::STATUS_PENDING,
            'attempts' => 0,
            'error_count' => 0,
            'skip_count' => 0,
            'last_error' => null,
        ]);

        return $clone;
    }

    public function storeFromWriteRequest(WriteRequest $writeRequest)
    {
        $existingTask = DocTask::query()
            ->where('title', $writeRequest->title)
            ->where('status', '!=', DocTaskService::STATUS_COMPLETED)
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
            'status' => DocTaskService::STATUS_PENDING,
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

        $pageTitle = DB::table('zetawiki.page')
            ->where('page_id', (int) $validated['page_id'])
            ->value('page_title');

        if (! is_string($pageTitle) || trim($pageTitle) === '') {
            return response()->json(['message' => '문서를 찾을 수 없습니다.'], 404);
        }

        $docTask = DocTask::create([
            'user_id' => auth()->id(),
            'user_name' => auth()->user()->name,
            'title' => str_replace('_', ' ', trim($pageTitle)),
            'request_type' => (string) $validated['request_type'],
            'content' => '',
            'status' => DocTaskService::STATUS_PENDING,
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
}
