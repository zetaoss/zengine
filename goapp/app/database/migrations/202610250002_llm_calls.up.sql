CREATE TABLE IF NOT EXISTS llm_calls (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  chain VARCHAR(64) NOT NULL,
  chain_version VARCHAR(16) NOT NULL,
  step VARCHAR(64) NOT NULL,
  model VARCHAR(128) NOT NULL,
  resolved_model VARCHAR(128) NOT NULL DEFAULT '',
  ok TINYINT(1) NOT NULL,
  error VARCHAR(300) NOT NULL DEFAULT '',
  prompt_tokens INT UNSIGNED NOT NULL DEFAULT 0,
  completion_tokens INT UNSIGNED NOT NULL DEFAULT 0,
  elapsed_ms INT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_llm_calls_model_ok (model, ok),
  KEY idx_llm_calls_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
