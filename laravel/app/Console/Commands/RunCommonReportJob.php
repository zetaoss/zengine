<?php
namespace App\Console\Commands;

use App\Jobs\CommonReportJob;
use App\Models\CommonReport;
use Illuminate\Console\Command;

class RunCommonReportJob extends Command
{
    protected $signature = 'job:run-common-report {items : Comma-separated list of words for the report items}';
    protected $description = 'Create and run the CommonReportJob with specified items';

    public function handle()
    {
        $itemsInput = $this->argument('items');
        $items = array_map('trim', explode(',', $itemsInput));

        if (empty($items)) {
            $this->error('At least one item must be provided.');
            return 1;
        }

        $report = new CommonReport();
        $report->state = 0;
        $report->user_id = 1;
        $report->save();

        foreach ($items as $word) {
            $report->items()->create([
                'name' => $word,
            ]);
        }

        $this->info("New report created with ID: {$report->id} and " . count($items) . " items.");

        try {
            CommonReportJob::dispatch($report->id);
            $this->info("CommonReportJob dispatched successfully for report ID: {$report->id}");
        } catch (\Exception $e) {
            $this->error("Failed to dispatch CommonReportJob: " . $e->getMessage());
            return 1;
        }

        return 0;
    }
}
