<?php

use MediaWiki\MediaWikiServices;
use ZetaExtension\Disambig\DisambigService;

require getenv('MW_INSTALL_PATH').'/maintenance/Maintenance.php';

class RebuildDisambigs extends Maintenance
{
    public function __construct()
    {
        parent::__construct();
        $this->addDescription('Rebuild all Disambig membership indexes and caches.');
    }

    public function execute(): void
    {
        $rows = MediaWikiServices::getInstance()
            ->getConnectionProvider()
            ->getReplicaDatabase()
            ->newSelectQueryBuilder()
            ->select(['page_id', 'page_title'])
            ->from('page')
            ->where(['page_namespace' => NS_DISAMBIG])
            ->orderBy('page_id')
            ->caller(__METHOD__)
            ->fetchResultSet();

        $successes = 0;
        $failures = [];
        foreach ($rows as $row) {
            $id = (int) $row->page_id;
            $title = str_replace('_', ' ', (string) $row->page_title);
            try {
                if (! is_array(DisambigService::ensureDisambig($id))) {
                    throw new RuntimeException('Disambig rebuild returned no data.');
                }
                $successes++;
                $this->output(sprintf("#%d %s\n", $id, $title));
            } catch (Throwable $e) {
                $failures[] = sprintf('#%d %s: %s', $id, $title, $e->getMessage());
            }
        }

        $this->output(sprintf("Rebuilt %d Disambig pages.\n", $successes));
        if ($failures !== []) {
            $this->fatalError("Failed Disambig pages:\n- ".implode("\n- ", $failures));
        }
    }
}

$maintClass = RebuildDisambigs::class;
require RUN_MAINTENANCE_IF_MAIN;
