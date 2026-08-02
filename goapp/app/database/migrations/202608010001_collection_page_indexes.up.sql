-- Extend Binder membership rows so missing pages can be indexed by normalized title.
SET @binder_page_namespace_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE() AND table_name = 'binder_pages' AND column_name = 'page_namespace'
);
SET @binder_page_namespace_sql := IF(
  @binder_page_namespace_exists = 0,
  'ALTER TABLE `binder_pages` ADD COLUMN `page_namespace` INT NULL AFTER `binder_id`',
  'SELECT 1'
);
PREPARE stmt FROM @binder_page_namespace_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @binder_page_title_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE() AND table_name = 'binder_pages' AND column_name = 'page_title'
);
SET @binder_page_title_sql := IF(
  @binder_page_title_exists = 0,
  'ALTER TABLE `binder_pages` ADD COLUMN `page_title` VARBINARY(255) NULL AFTER `page_namespace`',
  'SELECT 1'
);
PREPARE stmt FROM @binder_page_title_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `binder_pages` `bp`
JOIN `zetawiki`.`page` `p` ON `p`.`page_id` = `bp`.`page_id`
SET `bp`.`page_namespace` = `p`.`page_namespace`, `bp`.`page_title` = `p`.`page_title`
WHERE `bp`.`page_namespace` IS NULL OR `bp`.`page_title` IS NULL;

DELETE `bp` FROM `binder_pages` `bp`
LEFT JOIN `zetawiki`.`page` `p` ON `p`.`page_id` = `bp`.`page_id`
WHERE `bp`.`page_namespace` IS NULL OR `bp`.`page_title` IS NULL;

ALTER TABLE `binder_pages`
  MODIFY COLUMN `page_id` INT(10) UNSIGNED NULL,
  MODIFY COLUMN `page_namespace` INT NOT NULL,
  MODIFY COLUMN `page_title` VARBINARY(255) NOT NULL;

-- A Binder may mention more than one title that currently resolves to the same page.
-- Title identity, not the resolved page ID, is therefore the relation key.
SET @binder_pages_legacy_unique_exists := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE() AND table_name = 'binder_pages' AND index_name = 'page_id_2'
);
SET @binder_pages_legacy_unique_sql := IF(
  @binder_pages_legacy_unique_exists > 0,
  'ALTER TABLE `binder_pages` DROP INDEX `page_id_2`',
  'SELECT 1'
);
PREPARE stmt FROM @binder_pages_legacy_unique_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @binder_pages_pk_exists := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE() AND table_name = 'binder_pages' AND index_name = 'PRIMARY'
);
SET @binder_pages_pk_sql := IF(
  @binder_pages_pk_exists = 0,
  'ALTER TABLE `binder_pages` ADD PRIMARY KEY (`binder_id`, `page_namespace`, `page_title`)',
  'SELECT 1'
);
PREPARE stmt FROM @binder_pages_pk_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @binder_pages_title_index_exists := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE() AND table_name = 'binder_pages' AND index_name = 'binder_page_title'
);
SET @binder_pages_title_index_sql := IF(
  @binder_pages_title_index_exists = 0,
  'ALTER TABLE `binder_pages` ADD INDEX `binder_page_title` (`page_namespace`, `page_title`)',
  'SELECT 1'
);
PREPARE stmt FROM @binder_pages_title_index_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS `disambigs` (
  `id` INT(10) UNSIGNED NOT NULL,
  `cache` BLOB NULL,
  `entries` INT(10) UNSIGNED NOT NULL DEFAULT 0,
  `created_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=ascii COLLATE=ascii_general_ci ROW_FORMAT=COMPACT;

CREATE TABLE IF NOT EXISTS `disambig_pages` (
  `disambig_id` INT(10) UNSIGNED NOT NULL,
  `page_title` VARBINARY(255) NOT NULL,
  `page_id` INT(10) UNSIGNED NULL,
  PRIMARY KEY (`page_title`),
  UNIQUE KEY `disambig_page_id` (`page_id`),
  KEY `disambig_id` (`disambig_id`)
) ENGINE=InnoDB DEFAULT CHARSET=ascii COLLATE=ascii_general_ci ROW_FORMAT=COMPACT;

-- Normalize a pre-existing development schema that allowed non-main namespaces.
SET @disambig_page_namespace_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE() AND table_name = 'disambig_pages' AND column_name = 'page_namespace'
);
SET @disambig_non_main_delete_sql := IF(
  @disambig_page_namespace_exists > 0,
  'DELETE FROM `disambig_pages` WHERE `page_namespace` <> 0',
  'SELECT 1'
);
PREPARE stmt FROM @disambig_non_main_delete_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @disambig_page_namespace_sql := IF(
  @disambig_page_namespace_exists > 0,
  'ALTER TABLE `disambig_pages` DROP PRIMARY KEY, DROP COLUMN `page_namespace`, ADD PRIMARY KEY (`page_title`)',
  'SELECT 1'
);
PREPARE stmt FROM @disambig_page_namespace_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
