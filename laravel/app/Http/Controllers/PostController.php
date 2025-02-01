<?php

namespace App\Http\Controllers;

use App\Models\Post;
use Illuminate\Http\Request;
use Validator;

class PostController extends MyController
{
    public function recent()
    {
        return Post::orderBy('id', 'desc')->take(6)->get();
    }

    public function index()
    {
        return Post::orderBy('id', 'desc')->paginate(15);
    }

    public function show(Post $post)
    {
        $post->increment('hit');

        return $post;
    }

    public function store(Request $req)
    {
        $err = $this->validateRequest($req);
        if ($err !== false) {
            return $err;
        }
        $err = $this->shouldCreatable();
        if ($err !== false) {
            return $err;
        }
        $post = new Post;
        $post->cat = request('cat');
        $post->tags_str = '';
        $post->title = request('title');
        $post->body = request('body');
        $post->channel_id = 1;
        $post->user_id = $this->getUserID();
        $post->save();

        return ['status' => 'ok'];
    }

    public function update(Request $request, Post $post)
    {
        $err = $this->validateRequest($request);
        if ($err !== false) {
            return $err;
        }
        $err = $this->shouldEditable($post->user_id);
        if ($err !== false) {
            return $err;
        }
        $post->cat = request('cat');
        $post->tags_str = '';
        $post->title = request('title');
        $post->body = request('body');
        // $post->is_notice = request()->has('is_notice');
        $post->channel_id = 1;
        $post->user_id = $this->getUserID();
        $post->save();

        return ['status' => 'ok'];
    }

    public function destroy(Post $post)
    {
        $err = $this->shouldDeletable($post->user_id);
        if ($err !== false) {
            return $err;
        }
        $post->delete();

        return ['status' => 'ok'];
    }

    private function validateRequest(Request $request)
    {
        $validator = Validator::make($request->all(), [
            'cat' => 'required|in:질문,잡담,인사,기타',
            'title' => 'required|string|max:100',
            'body' => 'required|string|min:1,max:5000',
        ]);
        if ($validator->fails()) {
            return $this->newHTTPError(400, $validator->getMessageBag()->toArray());
        }

        return null;
    }
}
