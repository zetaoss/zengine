<?php

use ZetaExtension\Binder\BinderService;

require getenv('MW_INSTALL_PATH').'/maintenance/Maintenance.php';

class RebuildBinders extends Maintenance
{
    public function __construct()
    {
        parent::__construct();

        $this->addDescription('Rebuild all enabled binders.');
    }

    public function execute()
    {
        $binders = BinderService::listBinders();
        if ($binders === []) {
            $this->output("No binders found.\n");

            return;
        }

        $total = count($binders);
        $successes = 0;
        $failures = [];

        foreach ($binders as $index => $binder) {
            $id = (int) ($binder['id'] ?? 0);
            $title = (string) ($binder['title'] ?? '');
            $progress = sprintf('[%d/%d]', $index + 1, $total);

            try {
                $rebuilt = BinderService::ensureBinder($id);
                if (! is_array($rebuilt)) {
                    throw new RuntimeException('Binder rebuild returned no data.');
                }

                $successes++;
                $this->output(sprintf(
                    "%s #%d %s (docs=%d, links=%d, title_doc=%s)\n",
                    $progress,
                    $id,
                    $title,
                    (int) ($rebuilt['docs'] ?? 0),
                    (int) ($rebuilt['links'] ?? 0),
                    (string) ($rebuilt['title_doc'] ?? '')
                ));
            } catch (Throwable $e) {
                $failures[] = [
                    'id' => $id,
                    'title' => $title,
                    'error' => $e->getMessage(),
                ];
                $this->output(sprintf("%s #%d %s [fail] %s\n", $progress, $id, $title, $e->getMessage()));
            }
        }

        $this->output(sprintf("Rebuilt %d/%d binders.\n", $successes, $total));

        if ($failures !== []) {
            $this->output("Failed binders:\n");
            foreach ($failures as $failure) {
                $this->output(sprintf(
                    "- #%d %s: %s\n",
                    (int) ($failure['id'] ?? 0),
                    (string) ($failure['title'] ?? ''),
                    (string) ($failure['error'] ?? 'Unknown error')
                ));
            }
            $this->fatalError(sprintf('Failed to rebuild %d binders.', count($failures)));
        }
    }
}

$maintClass = RebuildBinders::class;
require RUN_MAINTENANCE_IF_MAIN;
