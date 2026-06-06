SET @aiedit_tasks_old_exists := (
  SELECT COUNT(*)
  FROM information_schema.tables
  WHERE table_schema = DATABASE()
    AND table_name = 'edit_tasks'
);
SET @aiedit_tasks_new_exists := (
  SELECT COUNT(*)
  FROM information_schema.tables
  WHERE table_schema = DATABASE()
    AND table_name = 'aiedit_tasks'
);
SET @aiedit_tasks_sql := IF(
  @aiedit_tasks_old_exists = 1 AND @aiedit_tasks_new_exists = 0,
  'RENAME TABLE edit_tasks TO aiedit_tasks',
  'SELECT 1'
);
PREPARE stmt FROM @aiedit_tasks_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @aiedit_prompts_old_exists := (
  SELECT COUNT(*)
  FROM information_schema.tables
  WHERE table_schema = DATABASE()
    AND table_name = 'editbot_prompts'
);
SET @aiedit_prompts_new_exists := (
  SELECT COUNT(*)
  FROM information_schema.tables
  WHERE table_schema = DATABASE()
    AND table_name = 'aiedit_prompts'
);
SET @aiedit_prompts_sql := IF(
  @aiedit_prompts_old_exists = 1 AND @aiedit_prompts_new_exists = 0,
  'RENAME TABLE editbot_prompts TO aiedit_prompts',
  'SELECT 1'
);
PREPARE stmt FROM @aiedit_prompts_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @aiedit_prompt_favorites_old_exists := (
  SELECT COUNT(*)
  FROM information_schema.tables
  WHERE table_schema = DATABASE()
    AND table_name = 'editbot_prompt_favorites'
);
SET @aiedit_prompt_favorites_new_exists := (
  SELECT COUNT(*)
  FROM information_schema.tables
  WHERE table_schema = DATABASE()
    AND table_name = 'aiedit_prompt_favorites'
);
SET @aiedit_prompt_favorites_sql := IF(
  @aiedit_prompt_favorites_old_exists = 1 AND @aiedit_prompt_favorites_new_exists = 0,
  'RENAME TABLE editbot_prompt_favorites TO aiedit_prompt_favorites',
  'SELECT 1'
);
PREPARE stmt FROM @aiedit_prompt_favorites_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
