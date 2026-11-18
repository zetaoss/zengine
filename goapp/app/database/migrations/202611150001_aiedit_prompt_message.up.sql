-- An AI edit prompt template's spec writes its messages as message.system and message.user, a block each,
-- instead of one template block of <message role="..."> sections. Specs are saved as the editor and
-- `tool migrate` wrote them: template: |- with its lines indented by two spaces, a system section, then a user
-- section. Every line indented by two gets two more, so each message's lines sit under its key; the section tags
-- become the keys. template_format is gone too: jinja2 was its only value.
UPDATE aiedit_prompts
SET spec = REPLACE(
  REPLACE(
    REPLACE(
      REPLACE(
        REPLACE(spec, '\n  ', '\n    '),
        '\ntemplate: |-\n    <message role="system">\n', '\nmessage:\n  system: |-\n'),
      '\ntemplate: |-\n    <message role="user">\n', '\nmessage:\n  user: |-\n'),
    '\n    </message>\n    <message role="user">\n', '\n  user: |-\n'),
  '\n    </message>', '')
WHERE spec LIKE '%\ntemplate: |-\n  <message role="%';

UPDATE aiedit_prompts
SET spec = REPLACE(spec, '\ntemplate_format: jinja2\n', '\n')
WHERE spec LIKE '%\ntemplate_format: jinja2\n%';
