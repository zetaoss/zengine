<?php

namespace App\Http\Controllers;

use App\Models\StatMwDaily;
use App\Models\StatMwHourly;
use App\Support\StatWindow;
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
        $to = Carbon::instance(StatWindow::hourlyEnd());
        $from = $to->copy()->subHours(47);

        $rows = StatMwHourly::query()
            ->select(array_merge(['timeslot'], self::NAMES))
            ->whereBetween('timeslot', [$from->toDateTimeString(), $to->toDateTimeString()])
            ->orderBy('timeslot')
            ->get();

        $timeslots = [];
        for ($cursor = $from->copy(); $cursor->lte($to); $cursor = $cursor->addHour()) {
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
        if (! in_array($days, [15, 90], true)) {
            abort(404);
        }

        $to = Carbon::instance(StatWindow::dailyEnd());
        $from = $to->copy()->subDays($days - 1)->startOfDay();

        $rows = StatMwDaily::query()
            ->select(array_merge(['timeslot'], self::NAMES))
            ->whereDate('timeslot', '>=', $from->toDateString())
            ->whereDate('timeslot', '<=', $to->toDateString())
            ->orderBy('timeslot')
            ->get();

        $timeslots = [];
        for ($cursor = $from->copy(); $cursor->lte($to); $cursor = $cursor->addDay()) {
            $timeslots[] = $cursor->toDateString();
        }
        $series = $this->emptySeries(count($timeslots));

        foreach ($rows as $row) {
            $timeslot = Carbon::parse((string) $row->timeslot)->toDateString();
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
