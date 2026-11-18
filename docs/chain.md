# Chains

goapp defines LLM work as chains, following LangChain; each step's prompt is written as a [Semantic Kernel prompt](https://learn.microsoft.com/en-us/semantic-kernel/concepts/prompts/yaml-schema) with a jinja2 template. Types: `goapp/prompts/chain.go`; the template dialect: `goapp/prompts/jinja`; built-in chains are YAML files in `goapp/prompts/chains/`, one per chain (for example `doc-eval.yaml`), embedded into the binary.

| Term | Example | Meaning | Elsewhere |
| :--- | :--- | :--- | :--- |
| Chain | `doc-eval` | Steps run in order; a step's output feeds later steps | LangChain `RunnableSequence` |
| Step | `doc-eval/rubric` | One LLM call: its messages, rendered and sent to a model, its answer read in a format | LangChain `prompt \| model \| parser`; SK prompt function |
| Message | `message.system`, `message.user` | A jinja2 template per role | SK `template` (jinja2 only, so no `template_format`), split by role |
| User inputs | `hint`, `texts`, `backlinks` | What the requester gives in the run form; they reach the template as they are | graph input schema (LangGraph) |
| Inputs | `title`, `body` | What the task running the chain fills | |
| Output variable | `rubrics`, `json` | The name a later step reads the answer under, and the format it is read in | ADK `output_key`; `response_format` |
| Version | `v2.3` | `v<major>.<minor>`, recorded with each call and result; a stored result of the same major is current | |

"System prompt" here is the system message; "user prompt" is not used.

## Built-in chains

| Chain | File | Run by | Model in the file |
| :--- | :--- | :--- | :--- |
| `doc-define` | `doc-define.yaml` | `doc-define-loop` ([doc-eval.md](doc-eval.md)) | `chains.docDefine` |
| `doc-eval` | `doc-eval.yaml` | `doc-eval-loop`, which also runs `doc-define` when a page has no definition ([doc-eval.md](doc-eval.md)) | `chains.docEval` |
| `doc-create` | `doc-create.yaml` | `ai-edit`, request type `create` | `chains.docCreate` |
| `doc-update` | `doc-update.yaml` | `ai-edit`, request type `update` | `chains.docUpdate` |

The model picked in the chains tab comes first; see [Screen and API](#screen-and-api) and [config.md](config.md#goapp-config-file). For `doc-create` and `doc-update` the requester may pick an AI edit prompt template (`aiedit_prompts`, see [Prompt templates](#prompt-templates)).

## Chain file

```yaml
name: doc-create
description: 새 문서의 위키텍스트를 작성한다.
version: v2.8
user_inputs:                    # the run form
  - name: hint
    description: 제목이 무엇을 가리키는지 가늠하게 하는 짧은 말
    ui: {placeholder: "예시: 제목이 '눈'이면 snow, '밤'이면 chestnut"}
  - name: texts
    json_schema: {type: array, items: {type: string}}
    ui: {variant: longText}
  - name: backlinks
    json_schema: {type: array, items: {type: string}}
inputs:                         # filled by the task
  - title
steps:
  - name: generate
    message:
      system: |-
        ...
      user: |-
        {% if texts %}
        <texts>
        {% for t in texts %}
        <text>
        {{ t }}
        </text>
        {% endfor %}
        </texts>
        {% endif %}
        ...

        ## 과제
        제목이 '{{ title }}'인 새 문서를 한국어 MediaWiki 위키텍스트로 쓴다.{% if hint %} 제목이 가리키는 뜻의 단서는 '{{ hint }}'이다.{% endif %}
        ...
```

Keys follow Semantic Kernel (snake_case); variable names are camelCase and start with a letter (`loop` is taken). The files are read at start with unknown fields rejected, and a missing file or step (`Chain.Step`) panics there.

### Variables

A variable's value has the shape of its `json_schema`, a closed subset of JSON Schema: a string (the default when it has none), a string with `enum` (a choice), a `boolean`, or an `array` of strings. `ui` tells the run form how to show it, with the names of [A2UI](https://github.com/a2ui-project/a2ui)'s `TextField`: `variant` `shortText` (one line, the default) or `longText` (several lines), and a `placeholder` for an empty field. Any other schema or `ui` is refused when the chain loads. A user input may also be `is_required` (a run without it, or with it blank, is refused; not for a boolean) and have a `default` (its value when the run form leaves it out; not for an array). A task input is a plain string written as its name alone (or with a `description`). In Go each variable has an `InputType` read from these (`string`, `text`, `choice`, `bool`, `textList`), which `GET /api/ai/chains` also sends as `type`.

### Steps

A step has a `name`, an optional `description` and its `message`; `output_variable` gives the `format` its answer is read in (`text` when omitted, or `json`) and a `name` when a later step reads it (doc-eval's `rubrics`). A step does not list its inputs: the variables its messages read are found when the chain loads, as LangChain finds a template's `input_variables`.

The doc-define template reads `{{ wikitext[:1000] }}`; doc-eval's judge reads the full wikitext, and the service checks quoted errors against that same text. Long documents therefore use more model context and tokens. Go keeps what the services need beyond the file: the structs a response is parsed into, values that check results (`DocEvalSafetyPolicies`), and the code between steps.

### Messages

A step's `message` has a `system` and a `user` key, each a jinja2 template written as a YAML block (`|-`). Either may be left out, but not both; they are sent system first. Each is parsed on its own (`prompts.ParseChat`), so a value can never move text into the other role, and an error names the key (`message.user: line 3, column 1: …`). A message that renders blank is not sent. Two messages of one role, or an assistant message, cannot be written; no chain needs them. Built-in steps name the page in the task sentence (`제목이 '{{ title }}'인 문서…`, doc-create's hint likewise) and put the data they read at the top of the user message, in tags under `## 입력` (`<body>…</body>`); an optional one is left out when empty, and a list holds one tag per item (`<texts>` holding a `<text>` each). The title is one line and is what the task is about, so it reads best in the sentence; `제목이 '…'인` also needs no particle chosen by the title's last letter. Each built-in step puts the task and output instructions after its inputs in the user message; its system message supplies the role and context.

## Template dialect

`goapp/prompts/jinja` renders a small dialect of Jinja2 that lives with wikitext.

| | Supported |
| :--- | :--- |
| Output | `{{ name }}`, `{{ loop.index }}`, `{{ name[start:end] }}` (by characters, negative bounds from the end); a list prints its items on lines of their own |
| Tags | `{% if %}` / `{% elif %}` / `{% else %}` / `{% endif %}` with `not`, `and`, `or`; `{% for item in list %}` / `{% endfor %}` with `loop.index`, `loop.index0`, `loop.first`, `loop.last`; `{% raw %}` / `{% endraw %}` |
| Other | `{# comments #}`, `-` whitespace markers (`{{- x -}}`, `{%- if x %}`) |
| Settings | `trim_blocks` and `lstrip_blocks` on: a line holding only a tag leaves no blank line. Nothing is escaped; values are inserted as they are and never parsed again |

Differences from Jinja2, and how wiki syntax gets through:

- `{{ … }}` is output only when its inside reads as an expression: it starts with a variable of the template (then the rest must parse), or it is a dotted ASCII name alone. Standard MediaWiki magic-word variables in uppercase, such as `{{PAGENAME}}`, stay as text. Anything else stays as text too, so wiki syntax passes: `{{풀기:사전링크}}`, `{{틀|a=1}}`, `{{#if:x}}`, `{{!}}`, and `{{틀|{{ title }}}}` keeps its outer braces around the title.
- `{{{` is always text, so wiki parameters pass with their names (`{{{1}}}`, `{{{title|기본}}}`); an output inside one still renders (`{{{1|{{ title }}}}}`).
- `{% … %}` is a tag only when a word follows `{%`; a word other than the tags above (`set`, `macro`, …) is an error.
- A name the template is not given is an error rather than empty, so a typo (`{{ titel }}`) is caught on load or save. Only the standard uppercase MediaWiki variable names are exempt; lowercase or unknown ASCII names remain errors.
- A blank string is false for `if`, as are `false` and an empty list.
- Filters, tests, comparisons, `set`, `macro`, `include` and `for … else` are not supported and are errors.

Where the dialect and Jinja2 overlap, the output is the same as Jinja2's with `trim_blocks` and `lstrip_blocks` (checked against Jinja2 3.1).

## Rules

`Chain.Compile` checks these at start (`mustLoad`), and in `chain_test.go`:

- Variable and output names are distinct; task inputs are plain strings.
- Every step's messages are jinja2, at least one of `system` and `user` is written, and they read only the chain's variables and the outputs of earlier steps; a `for` loops over an array.
- Every task input is read by some step; a user input need not be.
- `Format.Parse` strips a code fence around the whole response; `json` decodes into a struct, `text` trims. It checks the format only; the service checks what values mean.

## Running

`llmsvc.RunStep(ctx, model, chain, step, vars, out)` renders and runs; `RunMessages` runs messages rendered earlier (AI edit stores them with the task). Either calls bob `/aigate` and parses; the chain names each call in the call history (`llm_calls`, see [goapp.md](goapp.md)). A failed parse is retried once; each attempt parses into a fresh value, `out` is set only on success, and the error keeps at most 500 characters of the response. When the retry call itself fails, its error keeps the parse error as `previous attempt`. The service code runs the steps in order and the code between them.

`ai-edit` fills `title` and, for doc-update, `body` (the page's wikitext as it stands in the editor, else the saved page), takes the run form's values as they are (lists hold at most 20 items, and a value or item at most 100,000 characters), and renders the first step (`aiedit.Render`). Backlinks are a run form value: the run form's fetch button calls `GET /api/ai-edit/backlinks?title=`, which lists the titles of up to 10 main-namespace, non-redirect articles linking to the page (`services/wikisvc`), and the requester may edit the list before running.

## Prompt templates

An AI edit prompt template (`aiedit_prompts`) is a Semantic Kernel prompt document stored in `spec`, without `execution_settings`:

```yaml
name: doc-update-jmnote
description: 기존 문서의 위키텍스트를 고쳐 쓴다.
message:
  system: |-
    ...
  user: |-
    ...
input_variables:
  - name: title
  - name: body
  - name: texts
```

`input_variables` pick, by name, the variables of the request type's chain the template reads: its task inputs and any of its user inputs, each as the chain defines it (the run form shows the picked user inputs in the chain's order). Its `message` replaces the messages of the chain's first step and is checked by the chain's rules with only the picked variables (`Chain.WithTemplate`): a name the chain does not have, a repeated name, a variable read but not picked, or a task input not read is refused on save.

The model a template runs on is a setting of its own, outside the spec (`aiedit_prompts.model`; empty runs it on its chain's model). `name` and `description` are copied to the `title` and `description` columns for listing. A change to its messages or `input_variables` bumps the version (`v1.0` on creation; minor by default, major when picked); a change of name, description or model does not. Each run records the version as `prompt_version`.

Templates saved before specs (their `content`, `messages` and `user_inputs`) are converted by `tool migrate` after the SQL migrations (`aiedit.ConvertLegacyPrompts`). Stored messages keep their roles and order, each `if` and `each` message becoming a block around it (`{index}` is `{{ loop.index }}`), and one conditioned on a variable the chain no longer has (`request`) is left out; a template saved before messages has its content as the system message. `{name}` placeholders become `{{ name }}` (older names take their names now: `wikitext` is `body`, `backlinkTexts`, `backlinkTitles`, `readBacklinks` and `useBacklinks` are `backlinks`, `extraTexts` and `copiedTexts` are `texts`, in the editor's `-----name-----` headings too) and wiki markup that would read as template syntax goes inside `{% raw %}`. The user message then lists in tags the variables the template does not already read, under `## 입력` only when one of them is sent. A template that cannot be converted is logged and left without a spec, and cannot run until it is fixed.

Specs saved while a template was one `template` block of `<message role="...">` sections are rewritten to `message.system` and `message.user` by `tool migrate` after the legacy conversion (`aiedit.ConvertMessageSpecs`). It reads the YAML and the sections as the old parser did (either kind of quotes, blank lines between sections) and drops `template_format`, whose only value was jinja2. A spec it cannot carry over unchanged (an assistant message, two messages of one role, a user message before the system one) is logged and left as it is: the chains tab shows its error, and its owner edits its YAML there until it reads, when the form editor takes over.

## Chains and tasks

A chain says what to ask a model: its templates, formats, outputs and version, and which values the requester gives (user inputs) and the task fills (inputs). It does not pick targets, read the wiki, or store results; its model is picked per chain (`services/chainsvc`). A goapp task decides when to run a chain and on what: it picks the target, fills the inputs, runs the steps through its service, stores the result, handles failures, and decides from the chain version whether a stored result can be reused. See the table above for which task runs which chain.

## Version

`version` is `v<major>.<minor>`, recorded in full in `llm_calls` and with each stored result, so calls and results can be told apart by minor. A stored result is current when its major matches (`prompts.SameMajor`; a version written before minors, such as `v2`, counts as `v2.x`):

- Minor (`v2.0 → v2.1`): most edits. Stored definitions, evaluations and rubrics stay current.
- Major (`v2.1 → v3.0`): a change to the shape of the inputs, format or output, or an edit whose earlier results should all be redone. Every stored result of the chain becomes stale.

The model is outside the chain and does not change the version. Any change to the inputs or steps needs a bump: `TestSpecVersions` (`goapp/prompts/builtin_test.go`) pins each chain's version to a hash of its user inputs, inputs and steps, and fails until the version and its entry there are updated. It cannot tell a minor change from a major one; that is the author's call, and review's. See [doc-eval.md](doc-eval.md) for what doc-define and doc-eval reuse.

## Screen and API

The chains tab (`/tool/ai/chains`, sysop-only) is a table with a row per built-in chain and per AI edit prompt template: owner (`-` for a built-in), name, description, model, runs and last use. `GET /api/ai/chains` returns each chain as in its file plus `model`, `uses` and `used_at` (the count and latest time of the successful calls of its last step in `llm_calls`) and `template` (the chain as a prompt template in YAML, which a clone starts from).

- Built-in chains come first, on a gray background, most recently used first. `doc-create` and `doc-update` expand to the template editor, read-only, with the model picker and clone; the other chains expand to their version, inputs and steps (the variables each reads, its format and output, and its messages), with the model picker.
- Templates of `doc-create` and `doc-update` follow, most recently used first (`used_at`), then by use count; the requester's own come before the built-in chains, on light blue. A template that cannot run is marked with its error. A row expands to the template editor as it is (read-only), with edit and delete (owner only) and clone.
- The editor presents the template name, description, selected inputs and a `system` and a `user` message body beside the rendered preview; a body left empty is not written. There is no YAML editor: the form writes the spec. A spec the server cannot read shows its error and its YAML read-only, and cannot be edited or cloned (only deleted). On the right are temporary preview values for the task inputs (`title`, `body`) and selected user inputs, plus the rendered messages. `POST /api/ai-prompts/check` (`request_type`, `spec`, `preview_values`) returns these editor fields, the run form and rendered messages shortly after typing stops, or the error that would stop the save. Editing also offers the model (empty: the chain's) and, when the spec changed, a minor or major bump. Closing a row or leaving the page with unsaved edits asks first.
- Clone opens a new template at the top of the templates, being edited, with the source's spec under the name `<source>-<user name>` in lower case and the model the source runs on now, and creates it on save (`POST /api/ai-prompts` with `request_type`, `spec`, `model` and, for an edit, `id` and `bump`).
- `/tool/ai/prompts` and the old template pages redirect to the tab.

Each chain runs on one model, picked in the chains tab (`services/chainsvc`, `app_settings` `chain_models`). Until a sysop picks one, the `GOAPP_CONFIG` file decides (`chains.<chain>.model`, then `chains.common.model`), then `config.DefaultModel`. Several chains may share a model; the models tab lists each model's chains.

- `GET /api/ai/chains`: the built-in chains, each with its `model`
- `PUT /api/ai/chains/{name}/model`: `{"model": "provider/model"}`; an empty model removes the saved one, so the file decides again
