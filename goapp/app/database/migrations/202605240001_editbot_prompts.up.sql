CREATE TABLE IF NOT EXISTS editbot_prompts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  content LONGTEXT NOT NULL,
  use_count INT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY idx_editbot_prompts_title (title)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET @editbot_prompts_use_count_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'editbot_prompts'
    AND column_name = 'use_count'
);
SET @editbot_prompts_use_count_sql := IF(
  @editbot_prompts_use_count_exists = 0,
  'ALTER TABLE editbot_prompts ADD COLUMN use_count INT UNSIGNED NOT NULL DEFAULT 0 AFTER content',
  'SELECT 1'
);
PREPARE stmt FROM @editbot_prompts_use_count_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS editbot_prompt_favorites (
  user_id BIGINT UNSIGNED NOT NULL,
  prompt_id BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, prompt_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
