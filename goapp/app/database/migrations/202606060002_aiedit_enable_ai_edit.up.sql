SET @aiedit_tasks_enable_ai_edit_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'aiedit_tasks'
    AND column_name = 'enable_ai_edit'
);
SET @aiedit_tasks_enable_ai_edit_sql := IF(
  @aiedit_tasks_enable_ai_edit_exists = 0,
  'ALTER TABLE aiedit_tasks ADD COLUMN enable_ai_edit TINYINT(1) NOT NULL DEFAULT 0 AFTER phase',
  'SELECT 1'
);
PREPARE stmt FROM @aiedit_tasks_enable_ai_edit_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
