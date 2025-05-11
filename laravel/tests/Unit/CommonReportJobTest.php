<?php

namespace Tests\Unit;

use App\Jobs\CommonReportJob;
use App\Models\CommonReport;
use Tests\TestCase;

class CommonReportJobTest extends TestCase
{
    public function test_ok()
    {
        $items = ['Alice', 'Bob'];

        $report = new CommonReport;
        $report->phase = 'pending';
        $report->user_id = 1;
        $report->save();

        foreach ($items as $word) {
            $report->items()->create(['name' => $word]);
        }

        CommonReportJob::dispatch($report->id);

        $this->waitForComplete($report->id);

        $report = CommonReport::where('id', $report->id)->first();
        $this->assertEquals('succeeded', $report->phase);
    }

    private function waitForComplete($id, $timeout = 10)
    {
        $startTime = time();
        while (true) {
            $row = CommonReport::where('id', $id)->first();
            if ($row && $row->state == 'succeeded') {
                return;
            }

            if ((time() - $startTime) > $timeout) {
                $this->fail('Timeout');
            }

            usleep(500000); // 0.5s
        }
    }
}
