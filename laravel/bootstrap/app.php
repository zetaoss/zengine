<?php

use Illuminate\Foundation\Application;
use Illuminate\Foundation\Configuration\Exceptions;
use Illuminate\Foundation\Configuration\Middleware;

return Application::configure(basePath: dirname(__DIR__))
    ->withRouting(
        api: __DIR__.'/../routes/api.php',
        web: __DIR__.'/../routes/web.php',
        commands: __DIR__.'/../routes/console.php',
        health: '/up',
    )
    ->withMiddleware(function (Middleware $middleware) {
        $middleware->alias([
            'internal' => \App\Http\Middleware\InternalApiAuth::class,
            'mwauth' => \App\Http\Middleware\MwAuth::class,
        ]);

        $middleware->validateCsrfTokens(except: [
            'auth/deauthorize/facebook',
            'auth/deletion/facebook',
        ]);
    })
    ->withExceptions(function (Exceptions $exceptions) {
        //
    })->create();
