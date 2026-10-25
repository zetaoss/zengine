-- AI edit renders its chain on the server: a task records the template it ran (prompt_id, 0 for the chain's own
-- messages), the chain version, the variables it filled (inputs, JSON) and the rendered messages (llm_messages,
-- JSON) that the worker sends. llm_input keeps the messages joined as text.
ALTER TABLE aiedit_tasks
  ADD COLUMN IF NOT EXISTS prompt_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER request_type,
  ADD COLUMN IF NOT EXISTS prompt_version VARCHAR(16) NOT NULL DEFAULT '' AFTER prompt_id,
  ADD COLUMN IF NOT EXISTS chain_version VARCHAR(16) NOT NULL DEFAULT '' AFTER prompt_version,
  ADD COLUMN IF NOT EXISTS inputs LONGTEXT NULL AFTER phase,
  ADD COLUMN IF NOT EXISTS llm_messages LONGTEXT NULL AFTER inputs;

-- A template has a description, its own messages (JSON, in place of the chain step's; none: its content in place
-- of the unconditional ones), its own version (v<major>.<minor>, from v1.0, bumped when its messages change) and
-- its own model ('' runs it on its chain's model). used_at is when it was last run (unknown for runs before this,
-- which did not record their template).
ALTER TABLE aiedit_prompts
  ADD COLUMN IF NOT EXISTS version VARCHAR(16) NOT NULL DEFAULT 'v1.0' AFTER request_type,
  ADD COLUMN IF NOT EXISTS description VARCHAR(255) NOT NULL DEFAULT '' AFTER title,
  ADD COLUMN IF NOT EXISTS messages LONGTEXT NULL AFTER content,
  ADD COLUMN IF NOT EXISTS model VARCHAR(128) NOT NULL DEFAULT '' AFTER messages,
  ADD COLUMN IF NOT EXISTS used_at TIMESTAMP NULL AFTER use_count;

-- Template variables follow the chains' names: {제목} is {title}, {기존 문서 내용} (and {기존문서}) is {wikitext}.
UPDATE aiedit_prompts
SET content = REPLACE(REPLACE(REPLACE(content, '{제목}', '{title}'), '{기존 문서 내용}', '{wikitext}'), '{기존문서}', '{wikitext}')
WHERE content LIKE '%{제목}%' OR content LIKE '%{기존 문서 내용}%' OR content LIKE '%{기존문서}%';
