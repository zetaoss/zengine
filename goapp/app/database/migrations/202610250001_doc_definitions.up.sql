CREATE TABLE IF NOT EXISTS doc_definitions (
  page_id INT UNSIGNED NOT NULL,
  page_title VARCHAR(255) NOT NULL,
  revid BIGINT UNSIGNED NOT NULL,
  definition VARCHAR(500) NOT NULL,
  model VARCHAR(128) NOT NULL,
  chain_version VARCHAR(16) NOT NULL,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (page_id),
  KEY idx_doc_def_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- The definition an evaluation's rubrics were built from; rubrics are reused only for the same definition.
SET @doc_eval_definition_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'doc_evaluations'
    AND column_name = 'definition'
);
SET @doc_eval_definition_sql := IF(
  @doc_eval_definition_exists = 0,
  'ALTER TABLE doc_evaluations ADD COLUMN definition VARCHAR(500) NOT NULL DEFAULT \'\' AFTER revid',
  'SELECT 1'
);
PREPARE stmt FROM @doc_eval_definition_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
