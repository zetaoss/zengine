<?php

namespace App\Http\Controllers;

use App\Models\Comment;
use App\Services\AvatarService;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;

class CommentController extends Controller
{
    public function recent()
    {
        $rows = DB::connection('mwdb')->table('z_comment')
            ->select(
                'page.page_id',
                'page.page_title',
                'page.page_namespace',
                'z_comment.id',
                'z_comment.user_id',
                'z_comment.created',
                'z_comment.message'
            )
            ->join('page', 'z_comment.curid', 'page.page_id')
            ->orderBy('z_comment.created', 'desc')
            ->limit(10)
            ->get();

        $avatars = AvatarService::getAvatarsByIds($rows->pluck('user_id')->all());

        foreach ($rows as $row) {
            $uid = (int) $row->user_id;
            $row->avatar = $avatars[$uid] ?? null;
        }

        return $rows;
    }

    public function list($pageID)
    {
        $rows = DB::connection('mwdb')->table('z_comment')
            ->select('id', 'user_id', 'created', 'message')
            ->where('curid', (int) $pageID)
            ->orderBy('created', 'desc')
            ->get();

        $avatars = AvatarService::getAvatarsByIds($rows->pluck('user_id')->all());

        foreach ($rows as $row) {
            $uid = (int) $row->user_id;
            $row->avatar = $avatars[$uid] ?? null;
        }

        return $rows;
    }

    public function store(Request $request)
    {
        Gate::authorize('unblocked');

        $request->validate([
            'pageid' => 'required|integer|min:1',
            'message' => 'required|string|min:1|max:5000',
        ]);

        $userId = (int) auth()->id();
        $user = auth()->user();

        DB::connection('mwdb')->table('z_comment')->insert([
            'curid' => (int) $request->input('pageid'),
            'message' => (string) $request->input('message'),
            'user_id' => $userId,
            'name' => (string) ($user?->name ?? ''),
            'created' => now()->toDateTimeString(),
        ]);

        return ['ok' => true];
    }

    public function update(Comment $comment, Request $request)
    {
        Gate::authorize('owner', (int) $comment->user_id);

        $request->validate([
            'message' => 'required|string|min:1|max:5000',
        ]);

        $comment->message = (string) $request->input('message');
        $comment->save();

        return ['ok' => true];
    }

    public function destroy(Comment $comment)
    {
        Gate::authorize('ownerOrSysop', (int) $comment->user_id);

        $comment->delete();

        return ['ok' => true];
    }
}
