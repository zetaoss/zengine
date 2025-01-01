<?php
namespace App\Services;

use App\Models\Preview;

class PreviewService
{
    public static function getPreview($url)
    {
        $endpoint = getenv('EXTERNAL_API_ENDPOINT_V4');

        $url = self::sanitizeURL($url);
        if (!$url) {
            return ['status' => 'invalid'];
        }

        $url_md5 = md5($url);
        $preview = Preview::where('url_md5', $url_md5)->first();
        if ($preview) {
            return $preview;
        }

        $response = json_decode(file_get_contents("$endpoint/preview/v3/index.php?url=" . urlencode($url)));

        $preview = new Preview();
        $preview->url_md5 = $url_md5;
        $preview->url = $response->url ?? '';
        $preview->code = $response->code ?? 0;
        $preview->type = $response->type ?? '';
        $preview->title = $response->title ?? '';
        $preview->image = $response->image ?? '';
        $preview->description = $response->description ?? '';
        $preview->save();
        return $preview;
    }

    private static function sanitizeURL($url)
    {
        if (!str_starts_with($url, 'http')) {
            return false;
        }
        $url = preg_replace_callback('/[^\x20-\x7f]/', function ($match) {
            return urlencode($match[0]);
        }, $url);
        $url = str_replace('&amp;', '&', $url);
        if (filter_var($url, FILTER_VALIDATE_URL) === false) {
            return false;
        }
        return $url;
    }
}
