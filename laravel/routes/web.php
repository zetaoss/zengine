<?php

use App\Http\Controllers\SocialController;
use Illuminate\Support\Facades\Route;

Route::get('/auth/redirect/{provider}', [SocialController::class, 'redirect'])->where('provider', 'github|google');
Route::get('/auth/callback/{provider}', [SocialController::class, 'callback'])->where('provider', 'github|google');
