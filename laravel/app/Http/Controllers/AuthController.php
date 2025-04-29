<?php

namespace App\Http\Controllers;

use App\Services\AuthService;

class AuthController extends Controller
{
    public function logout()
    {
        $prefix = getenv('WG_COOKIE_PREFIX');
        setcookie($prefix.'Token', '', time() - 3600, '/');
        setcookie($prefix.'UserID', '', time() - 3600, '/');
        setcookie($prefix.'UserName', '', time() - 3600, '/');
        setcookie($prefix.'_session', '', time() - 3600, '/');

        return ['message' => 'logged out'];
    }

    public function me()
    {
        return AuthService::me();
    }
}
