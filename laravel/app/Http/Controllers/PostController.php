<?php

namespace App\Http\Controllers;

use App\Models\Post;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Gate;

class PostController extends Controller
{
    public function recent()
    {
        return Post::query()
            ->latest('id')
            ->take(6)
            ->get();
    }

    public function index()
    {
        return Post::query()
            ->latest('id')
            ->paginate(15);
    }

    public function show(Post $post)
    {
        $post->increment('hit');

        return $post;
    }

    public function store(Request $request)
    {
        Gate::authorize('unblocked');

        $validated = $request->validate([
            'cat' => 'required|in:질문,잡담,인사,기타',
            'title' => 'required|string|max:100',
            'body' => 'required|string|min:1|max:5000',
        ]);

        $post = Post::create([
            'cat' => (string) $validated['cat'],
            'title' => (string) $validated['title'],
            'body' => (string) $validated['body'],
            'tags_str' => '',
            'channel_id' => 1,
            'user_id' => (int) auth()->id(),
            'hit' => 0,
            'is_notice' => false,
        ]);

        return $post;
    }

    public function update(Request $request, Post $post)
    {
        Gate::authorize('owner', (int) $post->user_id);

        $validated = $request->validate([
            'cat' => 'required|in:질문,잡담,인사,기타',
            'title' => 'required|string|max:100',
            'body' => 'required|string|min:1|max:5000',
        ]);

        $post->update([
            'cat' => (string) $validated['cat'],
            'title' => (string) $validated['title'],
            'body' => (string) $validated['body'],
            'tags_str' => '',
            'channel_id' => 1,
        ]);

        return ['ok' => true];
    }

    public function destroy(Post $post)
    {
        Gate::authorize('ownerOrSysop', (int) $post->user_id);

        $post->delete();

        return ['ok' => true];
    }
}
