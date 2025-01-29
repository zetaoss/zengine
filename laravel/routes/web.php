<?php

use App\Http\Controllers\SocialController;
use Illuminate\Support\Facades\Route;

Route::prefix('auth')->group(function () {
    Route::get('/redirect/{provider}', [SocialController::class, 'redirect'])->where('provider', 'github|google');
    Route::get('/callback/{provider}', [SocialController::class, 'callback'])->where('provider', 'github|google');
});
