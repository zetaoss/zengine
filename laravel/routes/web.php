<?php

// web.php

use App\Http\Controllers\SocialController;
use App\Http\Controllers\SocialDeletionController;
use Illuminate\Support\Facades\Route;

Route::get('/auth/redirect/{provider}', [SocialController::class, 'redirect'])->where('provider', 'facebook|github|google')->middleware('throttle:30,1');
Route::get('/auth/callback/{provider}', [SocialController::class, 'callback'])->where('provider', 'facebook|github|google')->middleware('throttle:30,1');
Route::post('/auth/deletion/{provider}', [SocialDeletionController::class, 'create'])->where('provider', 'facebook')->middleware('throttle:30,1');
Route::get('/auth/deletion/{provider}/status/{code}', [SocialDeletionController::class, 'status'])->where('provider', 'facebook')->middleware('throttle:30,1');
