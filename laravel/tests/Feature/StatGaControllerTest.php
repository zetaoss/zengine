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
        $response->assertJsonPath('timeslots.0', '2026-03-14T17:00:00Z');
        $response->assertJsonPath('timeslots.47', '2026-03-16T16:00:00Z');
        $response->assertJsonPath('active_users.46', 11);
        $response->assertJsonPath('screen_page_views.46', 22);
        $response->assertJsonPath('sessions.46', 33);
        $response->assertJsonPath('active_users.47', null);
    } finally {
        Carbon::setTestNow();
        CarbonImmutable::setTestNow();
    }
});

it('anchors the hourly ga window to the current display cutoff', function () {
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
        $response->assertJsonPath('timeslots.0', '2026-03-15T19:00:00Z');
        $response->assertJsonPath('timeslots.47', '2026-03-17T18:00:00Z');
        $response->assertJsonPath('active_users.20', 11);
        $response->assertJsonPath('active_users.47', null);
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

        $response = $this->getJson('/api/stat/ga/daily/15');

        $response->assertOk();
        $response->assertJsonPath('timeslots.0', '2026-03-02');
        $response->assertJsonPath('timeslots.14', '2026-03-16');
        $response->assertJsonPath('active_users.14', 7);
        $response->assertJsonPath('screen_page_views.14', 14);
        $response->assertJsonPath('sessions.14', 21);
    } finally {
        Carbon::setTestNow();
        CarbonImmutable::setTestNow();
    }
});

it('anchors the daily ga window to the property timezone', function () {
    config()->set('services.google_analytics.timezone', 'America/Los_Angeles');

    $now = CarbonImmutable::parse('2026-03-24 01:00:00', 'UTC');

    Carbon::setTestNow($now);
    CarbonImmutable::setTestNow($now);

    try {
        StatGaDaily::query()->create([
            'timeslot' => '2026-03-14',
            'active_users' => 7,
            'screen_page_views' => 14,
            'sessions' => 21,
        ]);

        StatGaDaily::query()->create([
            'timeslot' => '2026-03-23',
            'active_users' => 8,
            'screen_page_views' => 16,
            'sessions' => 24,
        ]);

        $response = $this->getJson('/api/stat/ga/daily/15');

        $response->assertOk();
        $response->assertJsonPath('timeslots.0', '2026-03-09');
        $response->assertJsonPath('timeslots.14', '2026-03-23');
        $response->assertJsonPath('active_users.5', 7);
        $response->assertJsonPath('active_users.14', 8);
    } finally {
        Carbon::setTestNow();
        CarbonImmutable::setTestNow();
    }
});
