<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class EnsureNotBlocked
{
    public function handle(Request $request, Closure $next): Response
    {
        if ($this->isBlocked()) {
            abort(403, 'Access denied');
        }

        return $next($request);
    }

    private function isBlocked(): bool
    {
        $user_id = $_COOKIE['zetawikiUserID'] ?? false;
        if (! $user_id) {
            return true;
        }
        $row = DB::connection('mwdb')->table('ipblocks')->select('ipb_expiry')->where('ipb_user', '=', $user_id)->first();
        if ($row) {
            if ($row->ipb_expiry == 'infinity' || $row->ipb_expiry > date('YmdHis')) {
                return true;
            }
        }

        return false;
    }
}
