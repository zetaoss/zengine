-- The AI edit request type "edit" is renamed "update", matching the doc-update chain.
UPDATE aiedit_tasks SET request_type = 'update' WHERE request_type = 'edit';
UPDATE aiedit_prompts SET request_type = 'update' WHERE request_type = 'edit';
