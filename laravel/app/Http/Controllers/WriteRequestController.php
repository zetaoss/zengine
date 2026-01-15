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
        return $this->baseQuery()
            ->where('w.writer_id', '<', 0)
            ->orderByDesc('w.id')
            ->paginate(25);
    }

    public function indexTodoTop()
    {
        return $this->baseQuery()
            ->where('w.writer_id', '<', 0)
            ->orderByRaw('w.rate DESC, hit DESC, w.ref DESC, w.id DESC')
            ->paginate(25);
    }

    public function indexDone()
    {
        return $this->baseQuery()
            ->where('w.writer_id', '>', 0)
            ->orderByDesc('w.id')
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
        $wr->user_name = auth()->user()->name;
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

    private function baseQuery()
    {
        return DB::table('write_requests as w')
            ->select([
                'w.*',
                DB::raw('w.user_name as user_name'),
                DB::raw('(SELECT COALESCE(n.hit, 0) FROM not_matches n WHERE n.title = w.title LIMIT 1) as hit'),
            ]);
    }
}
