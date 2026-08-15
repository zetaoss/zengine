-- Persist Defender metrics collected by the Kubernetes statistics task.
-- The K8s tables are created lazily by the task, so leave environments where
-- they do not exist yet unchanged; AutoMigrate will create them with these
-- columns when collection starts.

SET @stat_k8s_hourly_fighting_ratio_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'stat_k8s_hourly'
    AND column_name = 'defender_fighting_ratio'
);
SET @stat_k8s_hourly_fighting_ratio_sql := IF(
  @stat_k8s_hourly_fighting_ratio_exists = 0
  AND EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = DATABASE() AND table_name = 'stat_k8s_hourly'
  ),
  'ALTER TABLE `stat_k8s_hourly` ADD COLUMN `defender_fighting_ratio` DOUBLE NOT NULL DEFAULT 0 AFTER `pvc_storage_capacity`',
  'SELECT 1'
);
PREPARE stmt FROM @stat_k8s_hourly_fighting_ratio_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @stat_k8s_hourly_max_level_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'stat_k8s_hourly'
    AND column_name = 'defender_max_level'
);
SET @stat_k8s_hourly_max_level_sql := IF(
  @stat_k8s_hourly_max_level_exists = 0
  AND EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = DATABASE() AND table_name = 'stat_k8s_hourly'
  ),
  'ALTER TABLE `stat_k8s_hourly` ADD COLUMN `defender_max_level` DOUBLE NOT NULL DEFAULT 0 AFTER `defender_fighting_ratio`',
  'SELECT 1'
);
PREPARE stmt FROM @stat_k8s_hourly_max_level_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @stat_k8s_daily_fighting_ratio_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'stat_k8s_daily'
    AND column_name = 'defender_fighting_ratio'
);
SET @stat_k8s_daily_fighting_ratio_sql := IF(
  @stat_k8s_daily_fighting_ratio_exists = 0
  AND EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = DATABASE() AND table_name = 'stat_k8s_daily'
  ),
  'ALTER TABLE `stat_k8s_daily` ADD COLUMN `defender_fighting_ratio` DOUBLE NOT NULL DEFAULT 0 AFTER `pvc_storage_capacity`',
  'SELECT 1'
);
PREPARE stmt FROM @stat_k8s_daily_fighting_ratio_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @stat_k8s_daily_max_level_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'stat_k8s_daily'
    AND column_name = 'defender_max_level'
);
SET @stat_k8s_daily_max_level_sql := IF(
  @stat_k8s_daily_max_level_exists = 0
  AND EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = DATABASE() AND table_name = 'stat_k8s_daily'
  ),
  'ALTER TABLE `stat_k8s_daily` ADD COLUMN `defender_max_level` DOUBLE NOT NULL DEFAULT 0 AFTER `defender_fighting_ratio`',
  'SELECT 1'
);
PREPARE stmt FROM @stat_k8s_daily_max_level_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
