<?php
namespace App\Http\Controllers\Box;

use App\Http\Controllers\Controller;
use App\Jobs\BoxJob;
use App\Models\RunBox;

class RunBoxController extends Controller
{
    // private function check($api, $lang)
    // {
    //     if (in_array($lang, ['bash', 'c', 'cpp', 'csharp', 'go', 'java', 'kotlin', 'lua', 'mysql', 'perl', 'php', 'powershell', 'python', 'r', 'ruby', 'sqlite3'])) {
    //         return null;
    //     }
    //     return "unsupported api($api) with lang($lang)";
    // }

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

    public function post($hash)
    {
        $validated = $request->validate([
            'pageId' => 'required|unique:posts|max:255',
            'req' => 'required',
        ]);
        $runBox = RunBox::where('hash', $hash)->first();
        if ($box) {
            return ['phase' => 9, 'message' => 'duplicated'];
        }
        $runBox = new RunBox;
        $box->hash = $hash;
        $box->step = 1;
        $box->pageId = request('pageId');
        $box->reqRun = request('reqRun');
        $box->lang = $lang;
        $box->source = $source;
        $box->save();
        // RunBoxJob::dispatch($box->id)->onQueue('runbox');
        // return ['step' => 1, 'message' => 'queued'];
    }

    private function postNotebook($name)
    {

    }

    public function put($api, $lang, $curid, $hash)
    {
        $err = $this->check($api, $lang);
        if (!is_null($err)) {
            return ['phase' => 9, 'message' => $err];
        }
        $box = Box::where('name', $name)->first();
        if (!$box) {
            return ['phase' => 9, 'message' => 'no'];
        }
        $box->phase = 1;
        $box->save();
        BoxJob::dispatch($box->id)->onQueue('runbox');
        return ['step' => 1, 'message' => 'queued'];
    }
}
