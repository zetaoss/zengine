<?php

namespace App\Http\Controllers;

use App\Models\CfAnalyticsDaily;
use App\Models\CfAnalyticsHourly;
use Illuminate\Support\Carbon;

class CfAnalyticsController extends Controller
{
    private const NAMES = [
        'uniq_uniques',
        'sum_requests',
        'sum_bytes',
        'sum_cachedBytes',
        'sum_browserMap',
    ];

    public function hourly(): array
    {
        $latestTimeslot = CfAnalyticsHourly::query()->max('timeslot');
        if (! $latestTimeslot) {
            return $this->emptySeriesResponse();
        }

        $to = Carbon::parse((string) $latestTimeslot, 'UTC')->startOfHour();
        $from = $to->copy()->subHours(23);

        $rows = CfAnalyticsHourly::query()
            ->select(['timeslot', 'name', 'value'])
            ->whereBetween('timeslot', [$from->toDateTimeString(), $to->toDateTimeString()])
            ->whereIn('name', self::NAMES)
            ->get();

        $timeslots = [];
        for ($cursor = $from->copy(); $cursor->lte($to); $cursor->addHour()) {
            $timeslot = $cursor->utc()->format('Y-m-d\TH:i:s\Z');
            $timeslots[] = $timeslot;
        }
        $series = $this->emptySeries(count($timeslots));

        foreach ($rows as $row) {
            $timeslot = Carbon::parse((string) $row->timeslot, 'UTC')->utc()->format('Y-m-d\TH:i:s\Z');
            $index = array_search($timeslot, $timeslots, true);
            if ($index === false) {
                continue;
            }
            $name = (string) $row->name;
            if (! array_key_exists($name, $series)) {
                continue;
            }
            $series[$name][$index] = $this->parseValueText((string) $row->value);
        }

        return ['timeslots' => $timeslots] + $series;
    }

    public function daily(int $days): array
    {
        if (! in_array($days, [7, 30], true)) {
            abort(404);
        }

        $lastDate = CfAnalyticsDaily::query()->max('timeslot');
        if (! $lastDate) {
            return $this->emptySeriesResponse();
        }

        $to = Carbon::parse((string) $lastDate)->startOfDay();
        $from = $to->copy()->subDays($days - 1)->startOfDay();

        $rows = CfAnalyticsDaily::query()
            ->select(['timeslot', 'name', 'value'])
            ->whereBetween('timeslot', [$from->toDateString(), $to->toDateString()])
            ->whereIn('name', self::NAMES)
            ->get();

        $timeslots = [];
        for ($cursor = $from->copy(); $cursor->lte($to); $cursor->addDay()) {
            $timeslots[] = $cursor->toDateString();
        }
        $series = $this->emptySeries(count($timeslots));

        foreach ($rows as $row) {
            $date = Carbon::parse((string) $row->timeslot)->toDateString();
            $index = array_search($date, $timeslots, true);
            if ($index === false) {
                continue;
            }
            $name = (string) $row->name;
            if (! array_key_exists($name, $series)) {
                continue;
            }
            $series[$name][$index] = $this->parseValueText((string) $row->value);
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

    private function parseValueText(string $value): mixed
    {
        $trimmed = trim($value);
        if ($trimmed === '') {
            return null;
        }

        $decoded = json_decode($trimmed, true);
        if (json_last_error() === JSON_ERROR_NONE) {
            return $decoded;
        }

        return null;
    }
}
