<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Symfony\Component\HttpFoundation\Response;

class EnsureNotBlocked
{
    public function handle(Request $request, Closure $next): Response
    {
        abort_if($this->isBlocked(), 403);

        return $next($request);
    }

    private function isBlocked(): bool
    {
        $prefix = config('app.wg_cookie_prefix');
        $user_id = $request->cookie("{$prefix}UserID");

        if (! $user_id) {
            return true;
        }

        $row = DB::connection('mwdb')->table('ipblocks')
            ->select('ipb_expiry')
            ->where('ipb_user', '=', $user_id)
            ->first();

        return $row && ($row->ipb_expiry === 'infinity' || $row->ipb_expiry > now()->format('YmdHis'));
    }
}
