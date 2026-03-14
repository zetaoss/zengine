<?php

namespace App\Http\Controllers;

use App\Models\StatDailyGa;
use App\Models\StatHourlyGa;
use Illuminate\Support\Carbon;

class StatGaController extends Controller
{
    public function hourly(): array
    {
        $latestTimeslot = StatHourlyGa::query()->max('timeslot');
        if (! $latestTimeslot) {
            return $this->emptySeriesResponse(StatHourlyGa::COLUMN_NAMES);
        }

        $to = Carbon::parse((string) $latestTimeslot, 'UTC')->startOfHour();
        $from = $to->copy()->subHours(23);

        $rows = StatHourlyGa::query()
            ->select(array_merge(['timeslot'], StatHourlyGa::COLUMN_NAMES))
            ->whereBetween('timeslot', [$from->toDateTimeString(), $to->toDateTimeString()])
            ->orderBy('timeslot')
            ->get();

        $timeslots = [];
        for ($cursor = $from->copy(); $cursor->lte($to); $cursor->addHour()) {
            $timeslots[] = $cursor->utc()->format('Y-m-d\TH:i:s\Z');
        }
        $series = $this->emptySeries(count($timeslots), StatHourlyGa::COLUMN_NAMES);

        foreach ($rows as $row) {
            $timeslot = Carbon::parse((string) $row->timeslot, 'UTC')->utc()->format('Y-m-d\TH:i:s\Z');
            $index = array_search($timeslot, $timeslots, true);
            if ($index === false) {
                continue;
            }

            foreach (StatHourlyGa::COLUMN_NAMES as $name) {
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

        $lastDate = StatDailyGa::query()->max('timeslot');
        if (! $lastDate) {
            return $this->emptySeriesResponse(StatDailyGa::COLUMN_NAMES);
        }

        $to = Carbon::parse((string) $lastDate)->startOfDay();
        $from = $to->copy()->subDays($days - 1)->startOfDay();

        $rows = StatDailyGa::query()
            ->select(array_merge(['timeslot'], StatDailyGa::COLUMN_NAMES))
            ->whereBetween('timeslot', [$from->toDateString(), $to->toDateString()])
            ->orderBy('timeslot')
            ->get();

        $timeslots = [];
        for ($cursor = $from->copy(); $cursor->lte($to); $cursor->addDay()) {
            $timeslots[] = $cursor->toDateString();
        }
        $series = $this->emptySeries(count($timeslots), StatDailyGa::COLUMN_NAMES);

        foreach ($rows as $row) {
            $timeslot = $row->timeslot->toDateString();
            $index = array_search($timeslot, $timeslots, true);
            if ($index === false) {
                continue;
            }

            foreach (StatDailyGa::COLUMN_NAMES as $name) {
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
