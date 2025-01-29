<?php
namespace App\Http\Controllers;

use App\Models\Runbox;

class RunboxController extends Controller
{
    public function get($page_id, $hash)
    {
        $r = Runbox::where('hash', $hash)->first();
        if (! $r) {
            return ['state' => 0];
        }
        return $r;
    }
}
