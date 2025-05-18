<?php

namespace App\Http\Controllers;

use App\Jobs\RunboxJob;
use App\Models\Runbox;
use Illuminate\Http\Request;

class RunboxController extends Controller
{
    public function show(string $hash)
    {
        $runbox = Runbox::where('hash', $hash)->first();
        if (! $runbox) {
            return ['phase' => 'none'];
        }

        return $runbox;
    }

    public function store(Request $request)
    {
        $validated = $request->validate([
            'hash' => 'required|string',
            'user_id' => 'required|integer',
            'page_id' => 'required|integer',
            'type' => 'required|string|in:lang,notebook',
            'payload' => 'required|array',
        ]);
        $runbox = Runbox::create([
            'hash' => $validated['hash'],
            'phase' => 'pending',
            'user_id' => 0, // TODO: 실제 사용자 ID 사용
            'page_id' => $validated['page_id'],
            'type' => $validated['type'],
            'payload' => $validated['payload'],
        ]);

        RunboxJob::dispatch($runbox->id);

        return response()->json(['phase' => 'pending'], 202);
    }
}
