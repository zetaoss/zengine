-- An AI edit prompt template is one Semantic Kernel style prompt document in YAML (spec): its name, description,
-- jinja2 template of messages, the chain variables it reads (input_variables) and its model (execution_settings).
-- title, description and model stay as columns for listing, copied from the spec on save. `goapp tool migrate`
-- converts the templates saved before (content, messages, user_inputs) into a spec after these statements; the
-- old columns are dropped by a later migration once every template has one.
ALTER TABLE aiedit_prompts
  ADD COLUMN IF NOT EXISTS spec LONGTEXT NULL AFTER version,
  MODIFY content LONGTEXT NULL;
