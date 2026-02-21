<?php

// web.php

use App\Http\Controllers\SocialController;
use App\Http\Controllers\SocialDetachController;
use Illuminate\Support\Facades\Route;

Route::get('/auth/redirect/{provider}', [SocialController::class, 'redirect'])->where('provider', 'facebook|github|google')->middleware('throttle:30,1');
Route::get('/auth/callback/{provider}', [SocialController::class, 'callback'])->where('provider', 'facebook|github|google')->middleware('throttle:30,1');
Route::post('/auth/deauthorize/{provider}', [SocialDetachController::class, 'deauthorize'])->where('provider', 'facebook')->middleware('throttle:30,1');
Route::post('/auth/deletion/{provider}', [SocialDetachController::class, 'deletion'])->where('provider', 'facebook')->middleware('throttle:30,1');
Route::get('/auth/deletion/{provider}/status/{code}', [SocialDetachController::class, 'deletionStatus'])->where('provider', 'facebook')->middleware('throttle:30,1');
