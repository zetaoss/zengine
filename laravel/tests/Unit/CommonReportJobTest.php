<?php
namespace Tests\Unit;

use App\Jobs\CommonReportJob;
use App\Models\CommonReport;
use Tests\TestCase;

class CommonReportJobTest extends TestCase
{
    public function testOk()
    {
        $items = ['Alice', 'Bob'];

        $report          = new CommonReport();
        $report->state   = 0;
        $report->user_id = 1;
        $report->save();

        foreach ($items as $word) {
            $report->items()->create(['name' => $word]);
        }

        $this->assertDatabaseHas('common_reports', [
            'id'      => $report->id,
            'state'   => 0,
            'user_id' => 1,
        ]);

        foreach ($items as $word) {
            $this->assertDatabaseHas('common_report_items', [
                'common_report_id' => $report->id,
                'name'             => $word,
            ]);
        }

        CommonReportJob::dispatch($report->id);

        sleep(1);

        $this->assertDatabaseHas('common_reports', [
            'id'    => $report->id,
            'state' => 1,
        ]);
    }
}
