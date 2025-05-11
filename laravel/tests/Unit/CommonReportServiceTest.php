<?php

namespace Tests\Unit;

use App\Jobs\CommonReportJob;
use App\Models\CommonReport;
use App\Models\CommonReportItem;
use App\Services\CommonReportService;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Queue;
use Tests\TestCase;

class CommonReportServiceTest extends TestCase
{
    protected CommonReportService $service;

    protected function setUp(): void
    {
        parent::setUp();
        $this->service = new CommonReportService;
    }

    public function test_it_creates_report_and_dispatches_job(): void
    {
        Queue::fake();

        $report = $this->service->create(['alpha', 'beta'], 1);

        $this->assertDatabaseHas('common_reports', ['id' => $report->id]);
        $this->assertEquals(2, CommonReportItem::where('report_id', $report->id)->count());

        Queue::assertPushed(CommonReportJob::class, function ($job) use ($report) {
            return $job->reportId === $report->id;
        });
    }

    public function test_it_clones_report(): void
    {
        Queue::fake();

        $original = CommonReport::factory()
            ->has(CommonReportItem::factory()->count(2))
            ->create();

        $clone = $this->service->clone($original, 99);

        $this->assertDatabaseHas('common_reports', ['id' => $clone->id, 'user_id' => 99]);
        $this->assertEquals(2, $clone->items()->count());

        Queue::assertPushed(CommonReportJob::class);
    }

    public function test_it_deletes_report(): void
    {
        $report = CommonReport::factory()->create();
        $this->service->delete($report);

        $this->assertModelMissing($report);
    }

    public function test_it_processes_report(): void
    {
        Http::fake([
            '*' => Http::response([
                'result' => [
                    'engines' => ['daum_blog', 'naver_blog'],
                    'values' => [
                        [10, 20],
                        [30, 40],
                    ],
                ],
            ]),
        ]);

        $report = CommonReport::factory()
            ->has(CommonReportItem::factory()->count(2))
            ->create();

        $this->service->processReport($report);

        $this->assertEquals('succeeded', $report->fresh()->phase);

        foreach ($report->items as $item) {
            $this->assertGreaterThan(0, $item->total);
        }
    }

    public function test_it_marks_timed_out_reports(): void
    {
        CommonReport::factory()->count(2)->create([
            'phase' => 'pending',
            'created_at' => now()->subMinutes(10),
        ]);

        $count = $this->service->markTimedOutReports(now()->subMinutes(5));

        $this->assertEquals(2, $count);

        $this->assertDatabaseHas('common_reports', ['phase' => 'failed']);
    }
}
