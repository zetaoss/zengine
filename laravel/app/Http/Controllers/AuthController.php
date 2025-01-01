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
        return ['message' => 'logged out'];
    }

    public function me()
    {
        return AuthService::me();
    }
}
