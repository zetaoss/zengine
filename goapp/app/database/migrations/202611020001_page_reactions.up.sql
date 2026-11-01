CREATE TABLE IF NOT EXISTS page_reactions (
  page_id INT UNSIGNED NOT NULL,
  cnt INT UNSIGNED NOT NULL DEFAULT 0,
  emoji_count TEXT COLLATE utf8mb4_unicode_ci NOT NULL,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (page_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS page_reaction_users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  page_id INT UNSIGNED NOT NULL,
  user_id INT UNSIGNED NOT NULL,
  emoji_code INT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_page_user_emoji (page_id, user_id, emoji_code),
  KEY idx_page_emoji (page_id, emoji_code),
  KEY idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
