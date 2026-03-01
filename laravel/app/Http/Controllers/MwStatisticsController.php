<?php

namespace App\Http\Controllers;

use App\Models\MwStatistics;
use Illuminate\Support\Carbon;

class MwStatisticsController extends Controller
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

    public function daily(int $days): array
    {
        if (! in_array($days, [7, 30], true)) {
            abort(404);
        }

        $lastDate = MwStatistics::query()->max('timeslot');
        if (! $lastDate) {
            return $this->emptySeriesResponse();
        }

        $to = Carbon::parse((string) $lastDate)->startOfDay();
        $from = $to->copy()->subDays($days - 1)->startOfDay();

        $rows = MwStatistics::query()
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
