<?php

use Illuminate\Support\Facades\DB;

it('forces sqlite in testing even when external env prefers mysql', function () {
    expect(app()->environment())->toBe('testing');
    expect(env('DB_CONNECTION'))->toBe('sqlite');
    expect(config('database.default'))->toBe('sqlite');
    expect(config('database.connections.sqlite.database'))->toBe(':memory:');

    DB::connection()->getPdo();

    expect(DB::connection()->getDriverName())->toBe('sqlite');
});
