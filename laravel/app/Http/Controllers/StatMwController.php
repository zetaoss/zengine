<?php

namespace App\Http\Controllers;

use App\Models\StatDailyMw;
use App\Models\StatHourlyMw;
use Illuminate\Support\Carbon;

class StatMwController extends Controller
{
    private const NAMES = [
        'pages',
        'articles',
        'edits',
        'images',
        'users',
        'activeusers',
        'admins',
        'jobs',
    ];

    public function hourly(): array
    {
        $latestTimeslot = StatHourlyMw::query()->max('timeslot');
        if (! $latestTimeslot) {
            return $this->emptySeriesResponse();
        }

        $to = Carbon::parse((string) $latestTimeslot, 'UTC')->startOfHour();
        $from = $to->copy()->subHours(23);

        $rows = StatHourlyMw::query()
            ->select(array_merge(['timeslot'], self::NAMES))
            ->whereBetween('timeslot', [$from->toDateTimeString(), $to->toDateTimeString()])
            ->orderBy('timeslot')
            ->get();

        $timeslots = [];
        for ($cursor = $from->copy(); $cursor->lte($to); $cursor->addHour()) {
            $timeslots[] = $cursor->utc()->format('Y-m-d\TH:i:s\Z');
        }
        $series = $this->emptySeries(count($timeslots));

        foreach ($rows as $row) {
            $timeslot = Carbon::parse((string) $row->timeslot, 'UTC')->utc()->format('Y-m-d\TH:i:s\Z');
            $index = array_search($timeslot, $timeslots, true);
            if ($index === false) {
                continue;
            }

            foreach (self::NAMES as $name) {
                $value = $row->{$name} ?? null;
                $series[$name][$index] = is_numeric($value) ? (float) $value : null;
            }
        }

        return ['timeslots' => $timeslots] + $series;
    }

    public function daily(int $days): array
    {
        if (! in_array($days, [7, 30], true)) {
            abort(404);
        }

        $lastDate = StatDailyMw::query()->max('timeslot');
        if (! $lastDate) {
            return $this->emptySeriesResponse();
        }

        $to = Carbon::parse((string) $lastDate)->startOfDay();
        $from = $to->copy()->subDays($days - 1)->startOfDay();

        $rows = StatDailyMw::query()
            ->select(array_merge(['timeslot'], self::NAMES))
            ->whereBetween('timeslot', [$from->toDateString(), $to->toDateString()])
            ->orderBy('timeslot')
            ->get();

        $timeslots = [];
        for ($cursor = $from->copy(); $cursor->lte($to); $cursor->addDay()) {
            $timeslots[] = $cursor->toDateString();
        }
        $series = $this->emptySeries(count($timeslots));

        foreach ($rows as $row) {
            $timeslot = $row->timeslot->toDateString();
            $index = array_search($timeslot, $timeslots, true);
            if ($index === false) {
                continue;
            }

            foreach (self::NAMES as $name) {
                $value = $row->{$name} ?? null;
                $series[$name][$index] = is_numeric($value) ? (float) $value : null;
            }
        }

        return ['timeslots' => $timeslots] + $series;
    }

    private function emptySeriesResponse(): array
    {
        return ['timeslots' => []] + $this->emptySeries(0);
    }

    private function emptySeries(int $size): array
    {
        $series = [];
        foreach (self::NAMES as $name) {
            $series[$name] = array_fill(0, $size, null);
        }

        return $series;
    }
}
