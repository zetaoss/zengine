<?php

namespace App\Http\Controllers;

use App\Models\StatGscDaily;
use App\Models\StatGscHourly;
use Illuminate\Support\Carbon;

class StatGscController extends Controller
{
    public function hourly(): array
    {
        $latestTimeslot = StatGscHourly::query()->max('timeslot');
        if (! $latestTimeslot) {
            return $this->emptySeriesResponse(StatGscHourly::COLUMN_NAMES);
        }

        $to = Carbon::parse((string) $latestTimeslot, 'UTC')->startOfHour();
        $from = $to->copy()->subHours(23);

        $rows = StatGscHourly::query()
            ->select(array_merge(['timeslot'], StatGscHourly::COLUMN_NAMES))
            ->whereBetween('timeslot', [$from->toDateTimeString(), $to->toDateTimeString()])
            ->orderBy('timeslot')
            ->get();

        $timeslots = [];
        for ($cursor = $from->copy(); $cursor->lte($to); $cursor->addHour()) {
            $timeslots[] = $cursor->utc()->format('Y-m-d\TH:i:s\Z');
        }
        $series = $this->emptySeries(count($timeslots), StatGscHourly::COLUMN_NAMES);

        foreach ($rows as $row) {
            $timeslot = Carbon::parse((string) $row->timeslot, 'UTC')->utc()->format('Y-m-d\TH:i:s\Z');
            $index = array_search($timeslot, $timeslots, true);
            if ($index === false) {
                continue;
            }

            foreach (StatGscHourly::COLUMN_NAMES as $name) {
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

        $lastDate = StatGscDaily::query()->max('timeslot');
        if (! $lastDate) {
            return $this->emptySeriesResponse(StatGscDaily::COLUMN_NAMES);
        }

        $to = Carbon::parse((string) $lastDate)->startOfDay();
        $from = $to->copy()->subDays($days - 1)->startOfDay();

        $rows = StatGscDaily::query()
            ->select(array_merge(['timeslot'], StatGscDaily::COLUMN_NAMES))
            ->whereBetween('timeslot', [$from->toDateString(), $to->toDateString()])
            ->orderBy('timeslot')
            ->get();

        $timeslots = [];
        for ($cursor = $from->copy(); $cursor->lte($to); $cursor->addDay()) {
            $timeslots[] = $cursor->toDateString();
        }
        $series = $this->emptySeries(count($timeslots), StatGscDaily::COLUMN_NAMES);

        foreach ($rows as $row) {
            $timeslot = $row->timeslot->toDateString();
            $index = array_search($timeslot, $timeslots, true);
            if ($index === false) {
                continue;
            }

            foreach (StatGscDaily::COLUMN_NAMES as $name) {
                $value = $row->{$name} ?? null;
                $series[$name][$index] = is_numeric($value) ? (float) $value : null;
            }
        }

        return ['timeslots' => $timeslots] + $series;
    }

    private function emptySeriesResponse(array $columnNames): array
    {
        return ['timeslots' => []] + $this->emptySeries(0, $columnNames);
    }

    private function emptySeries(int $size, array $columnNames): array
    {
        $series = [];
        foreach ($columnNames as $name) {
            $series[$name] = array_fill(0, $size, null);
        }

        return $series;
    }
}
