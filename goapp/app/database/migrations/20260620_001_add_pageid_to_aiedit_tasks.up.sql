-- Add page_id to aiedit_tasks table and rename title to page_title
ALTER TABLE `aiedit_tasks`
    ADD COLUMN `page_id` INT(10) UNSIGNED NOT NULL DEFAULT 0 AFTER `user_name`,
    CHANGE COLUMN `title` `page_title` VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    ADD INDEX `aiedit_tasks_page_id` (`page_id`);

-- Backfill page_id for completed tasks
UPDATE `aiedit_tasks`
SET `page_id` = IFNULL((
    SELECT `p`.`page_id`
    FROM `zetawiki`.`page` `p`
    WHERE `p`.`page_title` = REPLACE(`aiedit_tasks`.`page_title`, ' ', '_') AND `p`.`page_namespace` = 0
    LIMIT 1
), 0)
WHERE `phase` = 'Completed' AND `aiedit_tasks`.`page_id` = 0 AND `aiedit_tasks`.`page_title` != '';
