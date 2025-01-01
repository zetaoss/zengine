<?php
namespace App\Http\Controllers;

use App\Models\Comment;
use App\Services\UserService;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class CommentController extends MyController
{
    public function recent()
    {
        return DB::connection('mwdb')->table('z_comment')
            ->select('page.page_id', 'page.page_title', 'page.page_namespace', 'z_comment.id', 'z_comment.user_id', 'z_comment.created', 'z_comment.message')
            ->join('page', 'z_comment.curid', 'page.page_id')
            ->orderBy('z_comment.created', 'desc')->limit(10)->get()
            ->map(function ($row) {
                $row->avatar = UserService::getUserAvatar($row->user_id);
                return $row;
            });
    }

    public function list($pageID)
    {
        return DB::connection('mwdb')->table('z_comment')
            ->select('id', 'user_id', 'created', 'message')->where('curid', $pageID)->orderBy('created', 'desc')->get()
            ->map(function ($row) {
                $row->avatar = UserService::getUserAvatar($row->user_id);
                return $row;
            });
    }

    public function store(Request $request)
    {
        $request->validate([
            'pageid' => 'required|int|min:1',
            'message' => 'required|string|min:1,max:5000',
        ]);
        $err = $this->shouldCreatable();
        if ($err !== false) {
            return $err;
        }
        DB::connection('mwdb')->table('z_comment')->insert([
            'curid' => request('pageid'),
            'message' => request('message'),
            'user_id' => $this->getUserID(),
            'name' => $this->getUserName(),
            'created' => date('Y-m-d H:i:s'),
        ]);
        return ['status' => 'ok'];
    }

    public function update(Comment $comment, Request $request)
    {
        $request->validate([
            'message' => 'required|string|min:1,max:5000',
        ]);
        $err = $this->shouldEditable($comment->user_id);
        if ($err !== false) {
            return $err;
        }
        $comment->message = $request->message;
        $comment->save();
        return ['status' => 'ok'];
    }

    public function destroy(Comment $comment)
    {
        $err = $this->shouldDeletable($comment->user_id);
        if ($err !== false) {
            return $err;
        }
        $comment->delete();
        return ['status' => 'ok'];
    }

    public function catcomments($pageID)
    {
        $cats = DB::connection('mwdb')->select("SELECT cl.cl_to name, COUNT(c.id) cnt
        FROM z_comment c, categorylinks cl
        WHERE cl.cl_to IN (SELECT cl_to FROM categorylinks WHERE cl_from=?)
        AND c.curid = cl.cl_from
        AND cl.cl_type = 'page'
        GROUP BY cl_to
        ORDER BY cnt DESC", [$pageID]);
        return [
            'cats' => $cats,
            'comments' => (count($cats) == 0) ? [] : DB::connection('mwdb')->table('z_comment')
                ->select('page.page_title', 'z_comment.message', 'z_comment.user_id', 'z_comment.created')
                ->join('page', 'z_comment.curid', 'page.page_id')
                ->join('categorylinks', 'z_comment.curid', 'categorylinks.cl_from')
                ->where([['categorylinks.cl_type', 'page'], ['categorylinks.cl_to', $cats[0]->name]])->get()
                ->map(function ($row) {
                    $row->avatar = UserService::getUserAvatar($row->user_id);
                    return $row;
                }),
        ];
    }

    public function catcommentsCat($cat)
    {
        return [
            'comments' => DB::connection('mwdb')->table('z_comment')
                ->select('page.page_title', 'z_comment.message', 'z_comment.user_id', 'z_comment.created')
                ->join('page', 'z_comment.curid', 'page.page_id')
                ->join('categorylinks', 'z_comment.curid', 'categorylinks.cl_from')
                ->where([['categorylinks.cl_type', 'page'], ['categorylinks.cl_to', $cat]])->get()
                ->map(function ($row) {
                    $row->avatar = UserService::getUserAvatar($row->user_id);
                    return $row;
                }),
        ];
    }

}
