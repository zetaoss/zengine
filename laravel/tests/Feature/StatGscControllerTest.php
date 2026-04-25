<?php

use App\Models\StatGscDaily;
use Carbon\CarbonImmutable;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Carbon;

uses(RefreshDatabase::class);

it('anchors the daily gsc window to the search console timezone', function () {
    $now = CarbonImmutable::parse('2026-03-24 01:00:00', 'UTC');

    Carbon::setTestNow($now);
    CarbonImmutable::setTestNow($now);

    try {
        StatGscDaily::query()->create([
            'timeslot' => '2026-03-14',
            'clicks' => 7,
            'impressions' => 14,
            'ctr' => 50.0,
            'position' => 3.5,
        ]);

        StatGscDaily::query()->create([
            'timeslot' => '2026-03-23',
            'clicks' => 9,
            'impressions' => 18,
            'ctr' => 50.0,
            'position' => 2.5,
        ]);

        $response = $this->getJson('/api/stat/gsc/daily/15');

        $response->assertOk();
        $response->assertJsonPath('timeslots.0', '2026-03-09');
        $response->assertJsonPath('timeslots.14', '2026-03-23');
        $response->assertJsonPath('clicks.5', 7);
        $response->assertJsonPath('clicks.14', 9);
    } finally {
        Carbon::setTestNow();
        CarbonImmutable::setTestNow();
    }
});
