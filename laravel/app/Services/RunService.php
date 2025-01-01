<?php
namespace App\Services;

use Illuminate\Support\Facades\Http;

class RunService
{
    public static function lang($data): array
    {
        try {
            $response = Http::post("http://runbox/run/lang", $data);
            if ($response->successful()) {
                return [$response->json(), null];
            }
            return [null, "unsuccessful: " . $response->status()];
        } catch (\Exception $e) {
            return [null, "exception: " . $e->getMessage()];

        }
    }
}
