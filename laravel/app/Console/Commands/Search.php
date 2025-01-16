<?php
namespace App\Console\Commands;

use App\Services\SearchService;
use Illuminate\Console\Command;

class Search extends Command
{
    protected $signature   = 'app:search {type?} {word?}';
    protected $description = 'Run a search using the SearchService';

    public function handle()
    {
        $type = $this->argument('type');
        $word = $this->argument('word');

        $searchService = new SearchService();
        $validTypes    = $searchService->getTypes();

        if (! $type || ! in_array($type, $validTypes, true)) {
            $this->line('Invalid or missing required argument: type.');
            $this->line('Available types: ' . implode(', ', $validTypes));
            $this->line('Usage: php artisan app:search {type} {word}');
            return Command::FAILURE;
        }

        if (! $word) {
            $this->line('Missing required argument: word.');
            $this->line('Example: php artisan app:search ' . $type . ' hello');
            return Command::FAILURE;
        }

        $result = $searchService->search($type, $word);

        if (! $result['success']) {
            $this->error('Search failed: ' . ($result['error'] ?? 'Unknown error occurred.'));
            return Command::FAILURE;
        }

        $this->info("Search successful. Result: " . $result['data']);
        return Command::SUCCESS;
    }
}
