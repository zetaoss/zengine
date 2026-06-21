-- Add page_id to aiedit_tasks table.
SET @aiedit_tasks_page_id_exists := (
    SELECT COUNT(*)
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'aiedit_tasks'
      AND column_name = 'page_id'
);
SET @aiedit_tasks_page_id_sql := IF(
    @aiedit_tasks_page_id_exists = 0,
    'ALTER TABLE `aiedit_tasks` ADD COLUMN `page_id` INT(10) UNSIGNED NOT NULL DEFAULT 0 AFTER `user_name`',
    'SELECT 1'
);
PREPARE stmt FROM @aiedit_tasks_page_id_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Rename title to page_title when the old column is still present.
SET @aiedit_tasks_title_exists := (
    SELECT COUNT(*)
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'aiedit_tasks'
      AND column_name = 'title'
);
SET @aiedit_tasks_page_title_exists := (
    SELECT COUNT(*)
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'aiedit_tasks'
      AND column_name = 'page_title'
);
SET @aiedit_tasks_page_title_sql := IF(
    @aiedit_tasks_title_exists = 1 AND @aiedit_tasks_page_title_exists = 0,
    'ALTER TABLE `aiedit_tasks` CHANGE COLUMN `title` `page_title` VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL',
    'SELECT 1'
);
PREPARE stmt FROM @aiedit_tasks_page_title_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add the page_id index if it is missing.
SET @aiedit_tasks_page_id_index_exists := (
    SELECT COUNT(*)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'aiedit_tasks'
      AND index_name = 'aiedit_tasks_page_id'
);
SET @aiedit_tasks_page_id_index_sql := IF(
    @aiedit_tasks_page_id_index_exists = 0,
    'ALTER TABLE `aiedit_tasks` ADD INDEX `aiedit_tasks_page_id` (`page_id`)',
    'SELECT 1'
);
PREPARE stmt FROM @aiedit_tasks_page_id_index_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Backfill page_id for completed tasks
UPDATE `aiedit_tasks`
SET `page_id` = IFNULL((
    SELECT `p`.`page_id`
    FROM `zetawiki`.`page` `p`
    WHERE `p`.`page_title` = REPLACE(`aiedit_tasks`.`page_title`, ' ', '_') AND `p`.`page_namespace` = 0
    LIMIT 1
), 0)
WHERE `phase` = 'Completed' AND `aiedit_tasks`.`page_id` = 0 AND `aiedit_tasks`.`page_title` != '';
