<?php

use App\Models\StatGaDaily;
use App\Models\StatGaHourly;
use Carbon\CarbonImmutable;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Carbon;

uses(RefreshDatabase::class);

it('returns the hourly ga stat payload without hanging when using immutable display windows', function () {
    $now = CarbonImmutable::parse('2026-03-16 16:10:00', 'UTC');

    Carbon::setTestNow($now);
    CarbonImmutable::setTestNow($now);

    try {
        StatGaHourly::query()->create([
            'timeslot' => '2026-03-16 15:00:00',
            'active_users' => 11,
            'screen_page_views' => 22,
            'sessions' => 33,
        ]);

        $response = $this->getJson('/api/stat/ga/hourly');

        $response->assertOk();
        $response->assertJsonPath('timeslots.0', '2026-03-15T04:00:00Z');
        $response->assertJsonPath('timeslots.35', '2026-03-16T15:00:00Z');
        $response->assertJsonPath('active_users.35', 11);
        $response->assertJsonPath('screen_page_views.35', 22);
        $response->assertJsonPath('sessions.35', 33);
    } finally {
        Carbon::setTestNow();
        CarbonImmutable::setTestNow();
    }
});

it('anchors the hourly ga window to the latest stored timeslot', function () {
    $now = CarbonImmutable::parse('2026-03-17 18:20:00', 'UTC');

    Carbon::setTestNow($now);
    CarbonImmutable::setTestNow($now);

    try {
        StatGaHourly::query()->create([
            'timeslot' => '2026-03-16 15:00:00',
            'active_users' => 11,
            'screen_page_views' => 22,
            'sessions' => 33,
        ]);

        $response = $this->getJson('/api/stat/ga/hourly');

        $response->assertOk();
        $response->assertJsonPath('timeslots.35', '2026-03-16T15:00:00Z');
        $response->assertJsonPath('active_users.35', 11);
    } finally {
        Carbon::setTestNow();
        CarbonImmutable::setTestNow();
    }
});

it('returns the daily ga stat payload without hanging when using immutable display windows', function () {
    $now = CarbonImmutable::parse('2026-03-16 16:10:00', 'UTC');

    Carbon::setTestNow($now);
    CarbonImmutable::setTestNow($now);

    try {
        StatGaDaily::query()->create([
            'timeslot' => '2026-03-16',
            'active_users' => 7,
            'screen_page_views' => 14,
            'sessions' => 21,
        ]);

        $response = $this->getJson('/api/stat/ga/daily/7');

        $response->assertOk();
        $response->assertJsonPath('timeslots.0', '2026-03-10');
        $response->assertJsonPath('timeslots.6', '2026-03-16');
        $response->assertJsonPath('active_users.6', 7);
        $response->assertJsonPath('screen_page_views.6', 14);
        $response->assertJsonPath('sessions.6', 21);
    } finally {
        Carbon::setTestNow();
        CarbonImmutable::setTestNow();
    }
});
