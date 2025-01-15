<?php
namespace App\Console\Commands;

use App\Jobs\CommonReportJob;
use App\Models\CommonReport;
use Illuminate\Console\Command;

class NewCommonReport extends Command
{
    protected $signature   = 'app:new-common-report {items? : Comma-separated list of words for the report items}';
    protected $description = 'Create a CommonReport with items and dispatch a processing job.';

    public function handle()
    {
        $itemsInput = $this->argument('items');
        $items      = array_filter(array_map('trim', explode(',', $itemsInput)));
        if (empty($items)) {
            $this->info('At least one item must be provided.');
            $this->info('Usage: php artisan app:new-common-report Alice,Bob,Carol');
            return 1;
        }

        $report          = new CommonReport();
        $report->state   = 0;
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
