<?php

namespace App\Http\Controllers;

use App\Models\Post;
use App\Models\Reply;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Gate;

class ReplyController extends Controller
{
    public function index(Post $post)
    {
        return $post->replies()
            ->latest('created_at')
            ->get();
    }

    public function store(Post $post, Request $request)
    {
        Gate::authorize('unblocked');

        $validated = $request->validate([
            'body' => 'required|string|min:1|max:5000',
        ]);

        return $post->replies()->create([
            'body' => (string) $validated['body'],
            'user_id' => (int) auth()->id(),
        ]);
    }

    public function update(Post $post, Reply $reply, Request $request)
    {
        abort_unless((int) $reply->post_id === (int) $post->id, 404);

        Gate::authorize('owner', (int) $reply->user_id);

        $validated = $request->validate([
            'body' => 'required|string|min:1|max:5000',
        ]);

        $reply->update([
            'body' => (string) $validated['body'],
        ]);

        return ['ok' => true];
    }

    public function destroy(Post $post, Reply $reply)
    {
        abort_unless((int) $reply->post_id === (int) $post->id, 404);

        Gate::authorize('ownerOrSysop', (int) $reply->user_id);

        $reply->delete();

        return ['ok' => true];
    }
}
