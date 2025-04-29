<?php

namespace App\Http\Controllers;

use App\Services\AuthService;

class AuthController extends Controller
{
    public function logout()
    {
        setcookie('zetawikiToken', '', time() - 3600, '/');
        setcookie('zetawikiUserID', '', time() - 3600, '/');
        setcookie('zetawikiUserName', '', time() - 3600, '/');
        setcookie('zetawiki_session', '', time() - 3600, '/');

        // remove dot domain cookies (deprecated)
        $dot_domain = getenv('DOT_DOMAIN');
        if ($dot_domain) {
            setcookie('zetawikiToken', '', time() - 3600, '/', $dot_domain);
            setcookie('zetawikiUserID', '', time() - 3600, '/', $dot_domain);
            setcookie('zetawikiUserName', '', time() - 3600, '/', $dot_domain);
            setcookie('zetawiki_session', '', time() - 3600, '/', $dot_domain);
        }

        return ['message' => 'logged out'];
    }

    public function me()
    {
        return AuthService::me();
    }
}
