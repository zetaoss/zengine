SET @schema_name = DATABASE();

CREATE TABLE IF NOT EXISTS `stat_hourly_mw` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `timeslot` DATETIME NOT NULL COMMENT 'UTC hour',
    `pages` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `articles` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `edits` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `images` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `users` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `activeusers` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `admins` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `jobs` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `stat_hourly_mw_timeslot_unique` (`timeslot`),
    KEY `stat_hourly_mw_timeslot_index` (`timeslot`)
);

DROP PROCEDURE IF EXISTS rename_table_if_needed;
DELIMITER $$
CREATE PROCEDURE rename_table_if_needed(
    IN old_name VARCHAR(128),
    IN new_name VARCHAR(128)
)
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = @schema_name
          AND table_name = old_name
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = @schema_name
          AND table_name = new_name
    ) THEN
        SET @sql = CONCAT('RENAME TABLE `', old_name, '` TO `', new_name, '`');
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END$$
DELIMITER ;

DROP PROCEDURE IF EXISTS rename_index_if_needed;
DELIMITER $$
CREATE PROCEDURE rename_index_if_needed(
    IN table_name_in VARCHAR(128),
    IN old_index_name VARCHAR(128),
    IN new_index_name VARCHAR(128)
)
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = @schema_name
          AND table_name = table_name_in
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.statistics
        WHERE table_schema = @schema_name
          AND table_name = table_name_in
          AND index_name = old_index_name
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.statistics
        WHERE table_schema = @schema_name
          AND table_name = table_name_in
          AND index_name = new_index_name
    ) THEN
        SET @sql = CONCAT(
            'ALTER TABLE `', table_name_in,
            '` RENAME INDEX `', old_index_name,
            '` TO `', new_index_name, '`'
        );
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END$$
DELIMITER ;

CALL rename_table_if_needed('cf_analytics_daily', 'stat_daily_cf');
CALL rename_table_if_needed('cf_analytics_hourly', 'stat_hourly_cf');
CALL rename_table_if_needed('mw_statistics', 'stat_daily_mw');

CALL rename_index_if_needed(
    'stat_daily_cf',
    'cf_analytics_daily_timeslot_name_unique',
    'stat_daily_cf_timeslot_name_unique'
);
CALL rename_index_if_needed(
    'stat_daily_cf',
    'cf_analytics_daily_timeslot_index',
    'stat_daily_cf_timeslot_index'
);
CALL rename_index_if_needed(
    'stat_hourly_cf',
    'cf_analytics_hourly_timeslot_name_unique',
    'stat_hourly_cf_timeslot_name_unique'
);
CALL rename_index_if_needed(
    'stat_hourly_cf',
    'cf_analytics_hourly_timeslot_index',
    'stat_hourly_cf_timeslot_index'
);
CALL rename_index_if_needed(
    'stat_daily_mw',
    'mw_statistics_timeslot_unique',
    'stat_daily_mw_timeslot_unique'
);

DROP PROCEDURE IF EXISTS rename_index_if_needed;
DROP PROCEDURE IF EXISTS rename_table_if_needed;
