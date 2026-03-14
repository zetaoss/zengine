<?php

namespace App\Services\Stat;

use Illuminate\Database\Eloquent\Model;

class CollectGaPersistService
{
    public function persistRows(string $modelClass, array $rows): array
    {
        if (empty($rows)) {
            return ['inserted' => 0, 'updated' => 0, 'skipped' => 0];
        }

        $columnNames = $modelClass::COLUMN_NAMES;
        $timeslots = array_values(array_unique(array_map(static fn (array $row): string => (string) $row['timeslot'], $rows)));

        /** @var Model $modelClass */
        $existing = $modelClass::query()
            ->toBase()
            ->select(array_merge(['timeslot'], $columnNames))
            ->whereIn('timeslot', $timeslots)
            ->get();

        $existingByTimeslot = [];
        foreach ($existing as $item) {
            $existingByTimeslot[(string) $item->timeslot] = $this->metricSnapshot((array) $item, $columnNames);
        }

        $upsertRows = [];
        $inserted = 0;
        $updated = 0;
        $skipped = 0;

        foreach ($rows as $row) {
            $timeslot = (string) $row['timeslot'];
            $current = $this->metricSnapshot($row, $columnNames);
            $old = $existingByTimeslot[$timeslot] ?? null;

            if ($old === null) {
                $inserted++;
                $upsertRows[] = $row;

                continue;
            }

            if ($old !== $current) {
                $updated++;
                $upsertRows[] = $row;
            } else {
                $skipped++;
            }
        }

        if (! empty($upsertRows)) {
            $modelClass::query()->upsert($upsertRows, ['timeslot'], $columnNames);
        }

        return ['inserted' => $inserted, 'updated' => $updated, 'skipped' => $skipped];
    }

    private function metricSnapshot(array $row, array $columnNames): array
    {
        $snapshot = [];
        foreach ($columnNames as $columnName) {
            $snapshot[$columnName] = (int) ($row[$columnName] ?? 0);
        }

        return $snapshot;
    }
}
