-- Models are set per chain (app_settings chain_models, services/chainsvc) instead of in the evaluation settings.
-- The evaluation settings stored the model with every save, so a model equal to the built-in default
-- (groq/openai/gpt-oss-120b) is not taken as a sysop's choice: only other models move, and the chains then
-- follow the GOAPP_CONFIG file.
INSERT IGNORE INTO app_settings (setting_key, value_json)
SELECT 'chain_models', JSON_OBJECT()
FROM app_settings WHERE setting_key = 'doc_eval_settings';

UPDATE app_settings c
JOIN app_settings e ON e.setting_key = 'doc_eval_settings'
SET c.value_json = JSON_SET(c.value_json, '$."doc-eval"', JSON_UNQUOTE(JSON_EXTRACT(e.value_json, '$.model')))
WHERE c.setting_key = 'chain_models'
  AND JSON_UNQUOTE(JSON_EXTRACT(e.value_json, '$.model')) NOT IN ('', 'groq/openai/gpt-oss-120b')
  AND JSON_EXTRACT(c.value_json, '$."doc-eval"') IS NULL;

UPDATE app_settings c
JOIN app_settings e ON e.setting_key = 'doc_eval_settings'
SET c.value_json = JSON_SET(c.value_json, '$."doc-define"', JSON_UNQUOTE(JSON_EXTRACT(e.value_json, '$.define_model')))
WHERE c.setting_key = 'chain_models'
  AND JSON_UNQUOTE(JSON_EXTRACT(e.value_json, '$.define_model')) NOT IN ('', 'groq/openai/gpt-oss-120b')
  AND JSON_EXTRACT(c.value_json, '$."doc-define"') IS NULL;

UPDATE app_settings
SET value_json = JSON_REMOVE(JSON_REMOVE(value_json, '$.model'), '$.define_model')
WHERE setting_key = 'doc_eval_settings';
