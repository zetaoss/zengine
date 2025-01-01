<?php
namespace App\Http\Controllers;

use App\Services\PreviewService;

class PreviewController extends Controller
{
    public function show()
    {
        return PreviewService::getPreview(request('url', ''));
    }
}
