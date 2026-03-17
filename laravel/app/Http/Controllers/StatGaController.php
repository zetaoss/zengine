<?php

namespace App\Http\Controllers;

use App\Models\StatGaDaily;
use App\Models\StatGaHourly;
use App\Services\Stat\CollectGaApiService;
use App\Support\StatWindow;
use Carbon\CarbonImmutable;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

class StatGaController extends Controller
{
    private const HOURLY_TABLES = ['stat_hourly_ga', 'stat_ga_hourly'];

    private const DAILY_TABLES = ['stat_daily_ga', 'stat_ga_daily'];

    public function hourly(): array
    {
        $to = Carbon::instance(StatWindow::hourlyEnd());
        $from = $to->copy()->subHours(35);

        $rows = $this->loadRows(
            self::HOURLY_TABLES,
            array_merge(['timeslot'], StatGaHourly::COLUMN_NAMES),
            [$from->toDateTimeString(), $to->toDateTimeString()]
        );

        $timeslots = [];
        for ($cursor = $from->copy(); $cursor->lte($to); $cursor = $cursor->addHour()) {
            $timeslots[] = $cursor->utc()->format('Y-m-d\TH:i:s\Z');
        }
        $series = $this->emptySeries(count($timeslots), StatGaHourly::COLUMN_NAMES);

        foreach ($rows as $row) {
            $timeslot = Carbon::parse((string) $row->timeslot, 'UTC')->utc()->format('Y-m-d\TH:i:s\Z');
            $index = array_search($timeslot, $timeslots, true);
            if ($index === false) {
                continue;
            }

            foreach (StatGaHourly::COLUMN_NAMES as $name) {
                $value = $row->{$name} ?? null;
                $series[$name][$index] = is_numeric($value) ? (float) $value : null;
            }
        }

        return ['timeslots' => $timeslots] + $series;
    }

    public function daily(int $days, CollectGaApiService $api): array
    {
        if (! in_array($days, [7, 30], true)) {
            abort(404);
        }

        [, , , $timezone] = $api->resolveCredentials();
        $to = Carbon::instance(StatWindow::dailyEnd(CarbonImmutable::now($timezone)));
        $from = $to->copy()->subDays($days - 1)->startOfDay();

        $rows = $this->loadRows(
            self::DAILY_TABLES,
            array_merge(['timeslot'], StatGaDaily::COLUMN_NAMES),
            [$from->toDateString(), $to->toDateString()],
            true
        );

        $timeslots = [];
        for ($cursor = $from->copy(); $cursor->lte($to); $cursor = $cursor->addDay()) {
            $timeslots[] = $cursor->toDateString();
        }
        $series = $this->emptySeries(count($timeslots), StatGaDaily::COLUMN_NAMES);

        foreach ($rows as $row) {
            $timeslot = Carbon::parse((string) $row->timeslot)->toDateString();
            $index = array_search($timeslot, $timeslots, true);
            if ($index === false) {
                continue;
            }

            foreach (StatGaDaily::COLUMN_NAMES as $name) {
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

    private function loadRows(array $tables, array $columns, array $between, bool $dateOnly = false): array
    {
        $merged = [];

        foreach ($tables as $table) {
            if (! Schema::hasTable($table)) {
                continue;
            }

            $query = DB::table($table)
                ->select($columns)
                ->orderBy('timeslot');

            if ($dateOnly) {
                $query
                    ->whereDate('timeslot', '>=', $between[0])
                    ->whereDate('timeslot', '<=', $between[1]);
            } else {
                $query->whereBetween('timeslot', $between);
            }

            $rows = $query->get();

            foreach ($rows as $row) {
                $merged[(string) $row->timeslot] = $row;
            }
        }

        return array_values($merged);
    }
}
