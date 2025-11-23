<?php

namespace App\Http\Controllers;

use App\Models\WriteRequest;
use DB;

class WriteRequestController extends MyController
{
    public function count()
    {
        return DB::table('write_requests')
            ->selectRaw('COUNT(CASE WHEN writer_id>-1 THEN 1 ELSE NULL END) AS done')
            ->selectRaw('COUNT(CASE WHEN writer_id=-1 THEN 1 ELSE NULL END) AS todo')->first();
    }

    public function indexTodo()
    {
        return WriteRequest::where('writer_id', '<', 0)->orderBy('id', 'desc')->paginate(25);
    }

    public function indexTodoTop()
    {
        return WriteRequest::where('writer_id', '<', 0)->orderByRaw('rate DESC, hit DESC, ref DESC, id DESC')->paginate(25);
    }

    public function indexDone()
    {
        return WriteRequest::where('writer_id', '>', 0)->orderBy('id', 'desc')->paginate(25);
    }

    public function store()
    {
        $err = $this->shouldCreatable();
        if (! empty($err)) {
            return $ok;
        }
        $title = trim(request('title'));
        if ($title === '') {
            return $this->newHTTPError(422, '제목을 입력해주세요.');
        }
        $wr = new WriteRequest;
        $wr->user_id = $this->getUserID();
        $wr->title = $title;
        $wr->save();
    }

    public function destroy($id)
    {
        $row = WriteRequest::find($id);
        $err = $this->shouldDeletable($row->user_id);
        if ($err !== false) {
            return $err;
        }
        $row->delete();

        return ['status' => 'ok'];
    }
}
