<?php
namespace App\Http\Controllers;

use App\Jobs\BoxJob;
use App\Models\Box;

class BoxRunController extends Controller
{
    private function check($api, $lang)
    {
        if ($api == 'run') {
            if (in_array($lang, ['bash', 'c', 'cpp', 'csharp', 'go', 'java', 'kotlin', 'lua', 'mysql', 'perl', 'php', 'powershell', 'python', 'r', 'ruby', 'sqlite3'])) {
                return null;
            }
        } elseif ($api == 'notebook') {
            if (in_array($lang, ['python', 'r'])) {
                return null;
            }
        }
        return "unsupported api($api) with lang($lang)";
    }

    public function get($pageid, $hash)
    {
        $box = Box::where('hash', $hash)->where('curid', $pageid)->where('api', 'run')->first();
        if (!$box) {
            return ['phase' => 0, 'message' => 'not exist'];
        }
        if ($box->step == 1) {
            return ['phase' => 1, 'message' => 'queued'];
        }
        if ($box->step == 2) {
            return ['phase' => 2, 'message' => 'running'];
        }
        if ($box->step == 9) {
            return ['phase' => 9, 'message' => 'error'];
        }
        if ($box->phase != 3) {
            return ['phase' => -1, 'message' => 'unknown'];
        }
        return $box;
    }

    public function post($api, $name)
    {
        $box = Box::where('name', $name)->first();
        if ($box) {
            return ['phase' => 9, 'message' => 'duplicated'];
        }
        if ($api == 'notebook') {
            return postNotebook($name);
        }
        if ($api == 'run') {
            return postRun($name);
        }
    }

    private function postRun($name)
    {
        // $source = request('source');
        // if ($api == 'notebook') {
        //     $cells = [];
        //     foreach ($source as $cellsource) {
        //         $lines = [];
        //         foreach (explode("\n", $cellsource) as $line) {
        //             $lines[] = "$line\n";
        //         }
        //         $cells[] = $lines;
        //     }
        //     $source = $cells;
        // } elseif ($api == 'multi') {
        //     $meta = request('meta');
        //     $sources = [];
        //     $files = [];
        //     foreach ($source as $i => $cellsource) {
        //         $sources[] = $cellsource;
        //         $files[] = $meta['files'][$i];
        //     }
        //     $source = [
        //         'sources' => $sources,
        //         'files' => $files,
        //         'mainIdx' => $meta['mainIdx'],
        //     ];
        // }
        // $box = new Box;
        // $box->step = 1;
        // $box->curid = $curid;
        // $box->api = $api;
        // $box->hash = $hash;
        // $box->lang = $lang;
        // $box->source = $source;
        // $box->save();
        // BoxJob::dispatch($box->id)->onQueue('runbox');
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
