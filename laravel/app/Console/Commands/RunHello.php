<?php

namespace App\Console\Commands;

use App\Services\RunService;
use Illuminate\Console\Command;

class RunHello extends Command
{
    protected $signature = 'app:run-hello';
    protected $description = 'Command description';

    public function handle()
    {
        $this->lang();
    }

    private function lang()
    {
        $data = [
            "lang" => "bash",
            "files" => [
                ["name" => "greet.txt", "body" => "hello"],
                ["name" => "", "body" => "cat greet.txt"],
            ],
            "main" => 1,
        ];
        [$json, $err] = RunService::lang($data);
        if ($err !== null) {
            $this->error($err);
            return;
        }
        print_r($json);
        $this->info("ok");
    }
}
