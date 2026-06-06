CREATE TABLE IF NOT EXISTS app_settings (
  id INT NOT NULL AUTO_INCREMENT,
  setting_key VARCHAR(64) NOT NULL,
  value_json JSON NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uq_app_settings_setting_key (setting_key)
);
