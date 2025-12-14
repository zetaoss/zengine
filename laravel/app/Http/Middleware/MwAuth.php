<?php

namespace App\Http\Middleware;

use App\Auth\MwUser;
use Closure;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;

class MwAuth
{
    public function handle(Request $request, Closure $next, string $mode = 'required')
    {
        $ui = $this->fetchUserInfo($request);

        if (! $ui || ($ui['anon'] ?? false)) {
            return $this->unauthenticated($mode, $request, $next);
        }

        $id = (int) ($ui['id'] ?? 0);
        if ($id < 1) {
            return $this->unauthenticated($mode, $request, $next);
        }

        $user = new MwUser($ui);

        Auth::setUser($user);
        $request->setUserResolver(fn () => $user);

        return $next($request);
    }

    private function fetchUserInfo(Request $request): ?array
    {
        $cookieHeader = $request->header('Cookie');
        if (! $cookieHeader) {
            return null;
        }

        $url = config('app.url').'/w/api.php?action=query&meta=userinfo&uiprop=groups|blockinfo&format=json';
        $ch = curl_init($url);
        curl_setopt_array($ch, [
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_HTTPHEADER => [
                'Cookie: '.$cookieHeader,
                'Accept: application/json',
            ],
            CURLOPT_TIMEOUT => 5,
            CURLOPT_CONNECTTIMEOUT => 2,
        ]);

        $body = curl_exec($ch);
        if ($body === false) {
            curl_close($ch);

            return null;
        }

        $status = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);

        if ($status !== 200) {
            return null;
        }

        $data = json_decode($body, true);
        if (! is_array($data)) {
            return null;
        }

        $ui = $data['query']['userinfo'] ?? null;

        return is_array($ui) ? $ui : null;
    }

    private function unauthenticated(string $mode, Request $request, Closure $next)
    {
        if ($mode === 'maybe') {
            return $next($request);
        }

        return response()->json(['message' => 'Unauthenticated'], 401);
    }
}
