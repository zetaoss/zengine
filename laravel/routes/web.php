<?php

// web.php

use App\Http\Controllers\SocialController;
use Illuminate\Support\Facades\Route;

Route::get('/auth/redirect/{provider}', [SocialController::class, 'redirect'])->where('provider', 'github|google')->middleware('throttle:30,1');
Route::get('/auth/callback/{provider}', [SocialController::class, 'callback'])->where('provider', 'github|google')->middleware('throttle:30,1');
