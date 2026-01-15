<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;

class InternalApiAuth
{
    public function handle(Request $request, Closure $next)
    {
        $secretKey = (string) config('services.internal.secret_key');
        if ($secretKey === '') {
            return response()->json(['message' => 'Unauthorized'], 401);
        }

        $timestamp = (string) $request->header('X-Api-Timestamp', '');
        $signature = (string) $request->header('X-Api-Signature', '');
        if ($timestamp === '' || $signature === '') {
            return response()->json(['message' => 'Unauthorized'], 401);
        }

        if (! ctype_digit($timestamp) || abs(time() - (int) $timestamp) > 300) {
            return response()->json(['message' => 'Unauthorized'], 401);
        }

        $path = $request->getPathInfo();
        $query = $request->getQueryString() ?? '';
        $message = implode("\n", [$request->getMethod(), $path, $query, $timestamp]);
        $expected = hash_hmac('sha256', $message, $secretKey);

        if (! hash_equals($expected, $signature)) {
            return response()->json(['message' => 'Unauthorized'], 401);
        }

        return $next($request);
    }
}
