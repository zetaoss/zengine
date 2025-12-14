<?php

namespace App\Http\Middleware;

use App\Services\AuthService;
use Closure;
use Illuminate\Http\Request;

class MwAuth
{
    public function handle(Request $request, Closure $next)
    {
        $me = AuthService::me();

        if (! $me) {
            return response()->json(['message' => 'Unauthenticated'], 401);
        }

        $request->attributes->set('me', $me);

        return $next($request);
    }
}
