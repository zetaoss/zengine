<?php
namespace App\Http\Controllers;

use App\Jobs\RunboxJob;
use App\Models\Runbox;
use Illuminate\Http\Request;

class RunboxController extends Controller
{

    public function get($hash)
    {
        $r = Runbox::where('hash', $hash)->first();
        if (! $r) {
            return ['step' => 0];
        }
        return $r;
    }

    public function post(Request $request)
    {
        $v = $request->validate([
            'hash'    => 'required|string',
            'user_id' => 'required|int',
            'page_id' => 'required|int',
            'type'    => 'required|string',
            'payload' => 'required|array',
        ]);

        if (! in_array($v['type'], ['lang', 'notebook'])) {
            abort(404);
        }
        $runbox          = new Runbox();
        $runbox->hash    = $v['hash'];
        $runbox->step    = 1;
        $runbox->user_id = 0;
        $runbox->page_id = $v['page_id'];
        $runbox->type    = $v['type'];
        $runbox->payload = $v['payload'];
        $runbox->save();

        RunboxJob::dispatch($runbox->id);
        return response()->json(['message' => 'Accepted'], 202);
    }
}
