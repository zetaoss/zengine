<?php
namespace App\Http\Controllers;

use App\Jobs\RunboxJob;
use App\Models\Runbox;
use Illuminate\Http\Request;

class RunboxController extends Controller
{

    public function get($page_id, $hash)
    {
        $r = Runbox::where('hash', $hash)->first();
        if (! $r) {
            return ['state' => 0];
        }

        return $r;
    }

    public function post(Request $request, $pageId, $hash)
    {
        $payload = $request->validate([
            'lang'         => 'required|string',
            'files'        => 'required|array',
            'files.*.body' => 'required|string',
        ]);

        $runbox          = new Runbox();
        $runbox->type    = 'run';
        $runbox->state   = 0;
        $runbox->user_id = 0;
        $runbox->page_id = $pageId;
        $runbox->hash    = $hash;
        $runbox->payload = $payload;
        $runbox->save();

        RunboxJob::dispatch($runbox->id);
        return response()->json(['message' => 'Accepted'], 202);
    }
}
