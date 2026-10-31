-- The run form's copied texts are the user input texts, as the chains name it: a template's own run form
-- (user_inputs) and messages ({copiedTexts}, each: copiedTexts) follow. A template whose own run form already has
-- an input texts keeps copiedTexts as its own input, which still runs.
UPDATE aiedit_prompts
SET user_inputs = REGEXP_REPLACE(user_inputs, '\\bcopiedTexts\\b', 'texts'),
  messages = REGEXP_REPLACE(messages, '\\bcopiedTexts\\b', 'texts'),
  content = REGEXP_REPLACE(content, '\\bcopiedTexts\\b', 'texts')
WHERE (user_inputs REGEXP '\\bcopiedTexts\\b'
    OR messages REGEXP '\\bcopiedTexts\\b'
    OR content REGEXP '\\bcopiedTexts\\b')
  AND NOT COALESCE(user_inputs REGEXP '"name":\\s*"texts"', FALSE);
