<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;

class Sysop
{
    public function __construct(private readonly MwAuth $mwAuth) {}

    public function handle(Request $request, Closure $next)
    {
        return $this->mwAuth->handle($request, function (Request $request) use ($next) {
            $user = Auth::user();
            if (! $user) {
                return response()->json(['message' => 'Unauthenticated'], 401);
            }

            $groups = $user->groups ?? [];
            if (! is_array($groups) || ! in_array('sysop', $groups, true)) {
                return response()->json(['message' => 'Forbidden'], 403);
            }

            return $next($request);
        });
    }
}
