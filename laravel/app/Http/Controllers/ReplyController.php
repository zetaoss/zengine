<?php

namespace App\Http\Controllers;

use App\Models\Post;
use App\Models\Reply;
use Illuminate\Http\Request;
use Validator;

class ReplyController extends MyController
{
    public function index(Post $post)
    {
        return $post->replies()->get();
    }

    public function store(Post $post, Request $request)
    {
        $err = $this->validateRequest($request);
        if (! empty($err)) {
            return $err;
        }
        $err = $this->shouldCreatable();
        if (! empty($err)) {
            return $err;
        }

        return $post->createReply([
            'body' => $request->body,
            'user_id' => $this->getUserID(),
            'user_name' => $this->getUserName(),
        ]);
    }

    public function update(Post $post, Reply $reply, Request $request)
    {
        $err = $this->validateRequest($request);
        if (! empty($err)) {
            return $err;
        }
        $err = $this->shouldEditable($reply->user_id);
        if (! empty($err)) {
            return $err;
        }
        $reply->body = $request->body;
        $reply->save();

        return ['status' => 'ok'];
    }

    public function destroy(Post $post, Reply $reply)
    {
        $err = $this->shouldDeletable($reply->user_id);
        if (! empty($err)) {
            return $err;
        }
        $reply->delete();

        return ['status' => 'ok'];
    }

    private function validateRequest(Request $request)
    {
        $validator = Validator::make($request->all(), [
            'body' => 'required|string|max:5000',
        ]);
        if ($validator->fails()) {
            return $this->newHTTPError(400, 'bad_request', $validator->getMessageBag()->toArray());
        }

        return null;
    }
}
