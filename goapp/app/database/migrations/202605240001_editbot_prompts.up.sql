CREATE TABLE IF NOT EXISTS aiedit_prompts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  user_name VARCHAR(255) NOT NULL,
  title VARCHAR(255) NOT NULL,
  request_type VARCHAR(50) NOT NULL,
  content LONGTEXT NOT NULL,
  use_count INT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY idx_aiedit_prompts_title (title)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET @aiedit_prompts_user_id_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'aiedit_prompts'
    AND column_name = 'user_id'
);
SET @aiedit_prompts_user_id_sql := IF(
  @aiedit_prompts_user_id_exists = 0,
  'ALTER TABLE aiedit_prompts ADD COLUMN user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER id',
  'SELECT 1'
);
PREPARE stmt FROM @aiedit_prompts_user_id_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @aiedit_prompts_user_name_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'aiedit_prompts'
    AND column_name = 'user_name'
);
SET @aiedit_prompts_user_name_sql := IF(
  @aiedit_prompts_user_name_exists = 0,
  'ALTER TABLE aiedit_prompts ADD COLUMN user_name VARCHAR(255) NOT NULL DEFAULT \"\" AFTER user_id',
  'SELECT 1'
);
PREPARE stmt FROM @aiedit_prompts_user_name_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @aiedit_prompts_request_type_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'aiedit_prompts'
    AND column_name = 'request_type'
);
SET @aiedit_prompts_request_type_sql := IF(
  @aiedit_prompts_request_type_exists = 0,
  'ALTER TABLE aiedit_prompts ADD COLUMN request_type VARCHAR(50) NOT NULL DEFAULT \"create\" AFTER title',
  'SELECT 1'
);
PREPARE stmt FROM @aiedit_prompts_request_type_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @aiedit_prompts_use_count_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'aiedit_prompts'
    AND column_name = 'use_count'
);
SET @aiedit_prompts_use_count_sql := IF(
  @aiedit_prompts_use_count_exists = 0,
  'ALTER TABLE aiedit_prompts ADD COLUMN use_count INT UNSIGNED NOT NULL DEFAULT 0 AFTER content',
  'SELECT 1'
);
PREPARE stmt FROM @aiedit_prompts_use_count_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
