-- AI edit renders its chain on the server: a task records the template it ran (prompt_id, 0 for the chain's own
-- messages), the chain version, the variables it filled (inputs, JSON) and the rendered messages (llm_messages,
-- JSON) that the worker sends. llm_input keeps the messages joined as text.
ALTER TABLE aiedit_tasks
  ADD COLUMN IF NOT EXISTS prompt_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER request_type,
  ADD COLUMN IF NOT EXISTS chain_version VARCHAR(16) NOT NULL DEFAULT '' AFTER prompt_id,
  ADD COLUMN IF NOT EXISTS inputs LONGTEXT NULL AFTER phase,
  ADD COLUMN IF NOT EXISTS llm_messages LONGTEXT NULL AFTER inputs;

-- Template variables follow the chains' names: {제목} is {title}, {기존 문서 내용} (and {기존문서}) is {wikitext}.
UPDATE aiedit_prompts
SET content = REPLACE(REPLACE(REPLACE(content, '{제목}', '{title}'), '{기존 문서 내용}', '{wikitext}'), '{기존문서}', '{wikitext}')
WHERE content LIKE '%{제목}%' OR content LIKE '%{기존 문서 내용}%' OR content LIKE '%{기존문서}%';
