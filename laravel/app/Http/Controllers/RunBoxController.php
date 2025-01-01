<?php

namespace App\Http\Controllers;

use App\Models\RunBox;

class RunBoxController extends Controller
{
    public function get($pageid, $hash)
    {
        $runBox = RunBox::where('hash', $hash)->first();
        if (!$runBox) {
            return ['step' => 0, 'message' => 'not exist'];
        }
        if ($runBox->step == 1) {
            return ['step' => 1, 'message' => 'queued'];
        }
        if ($runBox->step == 2) {
            return ['step' => 2, 'message' => 'running'];
        }
        if ($runBox->step == 9) {
            return ['step' => 9, 'message' => 'error'];
        }
        if ($runBox->step != 3) {
            return ['step' => -1, 'message' => 'unknown'];
        }
        return $runBox;
    }
}
