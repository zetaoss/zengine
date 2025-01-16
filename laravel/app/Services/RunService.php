<?php
namespace App\Services;

use Illuminate\Support\Facades\Http;

class RunService
{
    public static function lang($data): array
    {
        $endpoint = getenv('RUNBOX_ENDPOINT');
        try {
            $response = Http::post($endpoint . "/run/lang", $data);
            if ($response->successful()) {
                return [$response->json(), null];
            }
            return [null, "unsuccessful: " . $response->status()];
        } catch (\Exception $e) {
            return [null, "exception: " . $e->getMessage()];

        }
    }
}
