<?php

namespace App\Http\Controllers;

use App\Models\PageComment;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Gate;

class PageCommentController extends Controller
{
    public function recent()
    {
        return PageComment::query()
            ->select([
                'page.page_title',
                'page_comments.id',
                'page_comments.page_id',
                'page_comments.user_id',
                'page_comments.user_name',
                'page_comments.created',
                'page_comments.message',
            ])
            ->join('zetawiki.page', 'page_comments.page_id', '=', 'page.page_id')
            ->orderByDesc('page_comments.created')
            ->limit(10)
            ->get();
    }

    public function list($pageID)
    {
        return PageComment::query()
            ->select([
                'id',
                'user_id',
                'user_name',
                'created',
                'message',
            ])
            ->where('page_id', $pageID)
            ->orderByDesc('id')
            ->get();
    }

    public function store(Request $request)
    {
        Gate::authorize('unblocked');

        $validated = $request->validate([
            'pageid' => ['required', 'integer', 'min:1'],
            'message' => ['required', 'string', 'min:1', 'max:5000'],
        ]);

        PageComment::create([
            'page_id' => $validated['pageid'],
            'message' => $validated['message'],
            'user_id' => auth()->id(),
            'user_name' => auth()->user()->name,
            'created' => now(),
        ]);

        return ['ok' => true];
    }

    public function update(PageComment $comment, Request $request)
    {
        Gate::authorize('owner', $comment->user_id);

        $validated = $request->validate([
            'message' => ['required', 'string', 'min:1', 'max:5000'],
        ]);

        $comment->update($validated);

        return ['ok' => true];
    }

    public function destroy(PageComment $comment)
    {
        Gate::authorize('ownerOrSysop', $comment->user_id);

        $comment->delete();

        return ['ok' => true];
    }
}
