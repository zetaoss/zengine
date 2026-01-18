<?php

namespace App\Http\Controllers;

use App\Models\Oneline;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Gate;

class OnelineController extends Controller
{
    public function recent()
    {
        return Oneline::orderByDesc('id')->take(15)->get();
    }

    public function index()
    {
        return Oneline::orderByDesc('id')->paginate(30);
    }

    public function store(Request $request)
    {
        Gate::authorize('unblocked');

        $validated = $request->validate([
            'message' => 'required|string|min:1|max:5000',
        ]);

        $oneline = Oneline::create([
            'message' => (string) $validated['message'],
            'user_id' => (int) auth()->id(),
            'user_name' => (string) auth()->user()->name,
            'created' => now()->toDateTimeString(),
        ]);

        return $oneline;
    }

    public function destroy(Oneline $oneline)
    {
        Gate::authorize('ownerOrSysop', (int) $oneline->user_id);

        $oneline->delete();

        return ['ok' => true];
    }
}
