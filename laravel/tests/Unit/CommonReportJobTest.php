<?php

namespace Tests\Unit;

use App\Jobs\CommonReportJob;
use App\Models\CommonReport;
use Illuminate\Support\Facades\Http;
use Tests\TestCase;

class CommonReportJobTest extends TestCase
{
    public function test_ok()
    {
        $items = ['Alice', 'Bob'];

        putenv('SEARCH_URL=http://example.test');
        Http::fake([
            '*' => Http::response([
                'result' => [
                    'engines' => ['daum_blog'],
                    'values' => [[1], [2]],
                ],
            ], 200),
        ]);

        $report = CommonReport::create([
            'user_id' => 1,
            'phase' => 'pending',
        ]);

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
            if ($row && $row->phase === 'succeeded') {
                return;
            }

            if ((time() - $startTime) > $timeout) {
                $this->fail('Timeout');
            }

            usleep(500000); // 0.5s
        }
    }
}
