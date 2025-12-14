<?php

namespace App\Http\Controllers;

use App\Models\WriteRequest;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;

class WriteRequestController extends Controller
{
    public function count()
    {
        return DB::table('write_requests')
            ->selectRaw('COUNT(CASE WHEN writer_id > -1 THEN 1 ELSE NULL END) AS done')
            ->selectRaw('COUNT(CASE WHEN writer_id = -1 THEN 1 ELSE NULL END) AS todo')
            ->first();
    }

    public function indexTodo()
    {
        return WriteRequest::where('writer_id', '<', 0)
            ->orderBy('id', 'desc')
            ->paginate(25);
    }

    public function indexTodoTop()
    {
        return WriteRequest::where('writer_id', '<', 0)
            ->orderByRaw('rate DESC, hit DESC, ref DESC, id DESC')
            ->paginate(25);
    }

    public function indexDone()
    {
        return WriteRequest::where('writer_id', '>', 0)
            ->orderBy('id', 'desc')
            ->paginate(25);
    }

    public function store(Request $request)
    {
        Gate::authorize('unblocked');

        $request->validate([
            'title' => 'required|string|min:1|max:255',
        ]);

        $title = trim((string) $request->input('title'));

        $wr = new WriteRequest;
        $wr->user_id = auth()->id();
        $wr->title = $title;
        $wr->save();

        return ['ok' => true, 'id' => $wr->id];
    }

    public function update(WriteRequest $writeRequest, Request $request)
    {
        Gate::authorize('owner', (int) $writeRequest->user_id);

        $request->validate([
            'title' => 'required|string|min:1|max:255',
        ]);

        $writeRequest->title = trim((string) $request->input('title'));
        $writeRequest->save();

        return ['ok' => true];
    }

    public function destroy(WriteRequest $writeRequest)
    {
        Gate::authorize('ownerOrSysop', (int) $writeRequest->user_id);

        $writeRequest->delete();

        return ['ok' => true];
    }
}
