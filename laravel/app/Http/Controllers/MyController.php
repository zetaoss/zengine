<?php

namespace App\Http\Controllers;

use App\Services\AuthService;
use Illuminate\Foundation\Auth\Access\AuthorizesRequests;
use Illuminate\Foundation\Bus\DispatchesJobs;
use Illuminate\Foundation\Validation\ValidatesRequests;

class MyController extends Controller
{
    use AuthorizesRequests, DispatchesJobs, ValidatesRequests;

    private static $me = null;

    protected function getMe()
    {
        if (is_null(self::$me)) {
            self::$me = AuthService::me();
        }

        return self::$me;
    }

    public function getUserID()
    {
        return $this->getMe()['avatar']['id'] ?? 0;
    }

    public function getUserName()
    {
        return $this->getMe()['avatar']['name'] ?? '';
    }

    private function isAdmin()
    {
        return in_array('Administrator', $this->getMe()['group'] ?? []);
    }

    public function shouldCreatable()
    {
        return $this->checkPermssion(0, false, 'not creatable');
    }

    public function shouldEditable($writerID)
    {
        return $this->checkPermssion($writerID, false, 'not editable');
    }

    public function shouldDeletable($writerID)
    {
        return $this->checkPermssion($writerID, true, 'not deletable');
    }

    private function checkPermssion($writerID, $allowAdmin, $reason = '-')
    {
        $me = $this->getMe();
        if (! $me) {
            return $this->newHTTPError(403, 'unauthorized: not logged in');
        }
        if ($writerID == 0 || $writerID == $this->getUserID() || ($allowAdmin && $this->isAdmin())) {
            return false;
        }

        return $this->newHTTPError(403, "unauthorized: $reason");
    }

    public function newHTTPError($code, $error)
    {
        return response()->json(['status' => 'error', 'error' => $error], $code);
    }
}
