<?php

namespace App\Http\Controllers;

use App\Models\Oneline;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Gate;

class OnelineController extends Controller
{
    public function recent()
    {
        return Oneline::orderBy('id', 'desc')->take(15)->get();
    }

    public function index()
    {
        return Oneline::query()
            ->orderByDesc('id')
            ->paginate(30);
    }

    public function show(Oneline $oneline)
    {
        //
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

        return Oneline::query()
            ->where('id', (int) $oneline->id)
            ->firstOrFail();
    }

    public function update(Request $request, Oneline $oneline)
    {
        //
    }

    public function destroy(Oneline $oneline)
    {
        Gate::authorize('ownerOrSysop', (int) $oneline->user_id);

        $oneline->delete();

        return ['ok' => true];
    }
}
