<?php

namespace App\Http\Controllers;

use App\Jobs\RunboxJob;
use App\Models\Runbox;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class RunboxController extends Controller
{
    public function store(Request $request): JsonResponse
    {
        $v = $request->validate([
            'hash' => 'required|string',
            'user_id' => 'required|int',
            'page_id' => 'required|int',
            'type' => 'required|string',
            'payload' => 'required|array',
        ]);

        if (! in_array($v['type'], ['lang', 'notebook'])) {
            abort(400, 'Invalid type');
        }

        $runbox = new Runbox;
        $runbox->hash = $v['hash'];
        $runbox->phase = 'pending';
        $runbox->user_id = $v['user_id'];
        $runbox->page_id = $v['page_id'];
        $runbox->type = $v['type'];
        $runbox->payload = $v['payload'];
        $runbox->save();

        RunboxJob::dispatch($runbox->id);

        return response()->json(['message' => 'Accepted'], 202);
    }

    public function show(string $hash): JsonResponse
    {
        $runbox = Runbox::where('hash', $hash)->first();

        if (! $runbox) {
            return response()->json(['phase' => 'none']);
        }

        return response()->json($runbox);
    }

    public function destroy(string $hash): JsonResponse
    {
        $runbox = Runbox::where('hash', $hash)->firstOrFail();
        $runbox->delete();

        return response()->json(['message' => 'Deleted']);
    }

    public function rerun(string $hash): JsonResponse
    {
        $runbox = Runbox::where('hash', $hash)->firstOrFail();
        $runbox->phase = 'pending';
        $runbox->save();

        RunboxJob::dispatch($runbox->id);

        return response()->json(['message' => 'Rerun dispatched'], 202);
    }
}
