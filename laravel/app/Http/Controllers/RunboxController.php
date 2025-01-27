<?php
namespace App\Http\Controllers;

use App\Models\Runbox;

class RunBoxController extends Controller
{
    public function get($hash)
    {
        $r = Runbox::where('hash', $hash)->first();
        if (! $r) {
            return ['state' => 0];
        }
        if ($r->state == 1) {
            return ['state' => 1];
        }
        if ($r->state == 2) {
            return ['state' => 2];
        }
        if ($r->state == -1) {
            return ['state' => -1];
        }
        return $r;
    }
}
