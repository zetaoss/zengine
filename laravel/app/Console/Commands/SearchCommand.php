<?php

namespace App\Console\Commands;

use App\Services\SearchService;
use Illuminate\Console\Command;

class SearchCommand extends Command
{
    protected $signature = 'search {type} {word}';
    protected $description = 'Run a search using the SearchService';

    public function handle()
    {
        $type = $this->argument('type');
        $word = $this->argument('word');

        try {
            $searchService = new SearchService();
            $result = $searchService->search($type, $word);

            if ($result === -1) {
                $this->error('Search failed: Unable to fetch results.');
            } elseif ($result === -2) {
                $this->error('Search failed: Response path not found.');
            } else {
                $this->info("Search successful. Result: $result");
            }
        } catch (\Exception $e) {
            $this->error('Error: ' . $e->getMessage());
        }
    }
}
