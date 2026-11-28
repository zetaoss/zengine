# Chains

goapp defines LLM work as chains, following LangChain; each step's prompt is written as a [Semantic Kernel prompt](https://learn.microsoft.com/en-us/semantic-kernel/concepts/prompts/yaml-schema) with a jinja2 template. Types: `goapp/prompts/chain.go`; the template dialect: `goapp/prompts/jinja`; built-in chains are YAML files in `goapp/prompts/chains/`, one per chain (for example `doc-eval.yaml`), embedded into the binary.

| Term | Example | Meaning | Elsewhere |
| :--- | :--- | :--- | :--- |
| Chain | `doc-eval` | Steps run in order; a step's output feeds later steps | LangChain `RunnableSequence` |
| Step | `doc-eval/rubric` | One LLM call: its messages, rendered and sent to a model, its answer read in a format | LangChain `prompt \| model \| parser`; SK prompt function |
| Message | `message.system`, `message.user` | A jinja2 template per role | SK `template` (jinja2 only, so no `template_format`), split by role |
| Inputs | `title`, `wikitext` | What the task running the chain fills | |
| Output variable | `rubrics`, `json` | The name a later step reads the answer under, and the format it is read in | ADK `output_key`; `response_format` |
| Version | `v2.3` | `v<major>.<minor>`, recorded with each call and result; a stored result of the same major is current | |

"System prompt" here is the system message; "user prompt" is not used.

## Built-in chains

| Chain | File | Run by | Model in the file |
| :--- | :--- | :--- | :--- |
| `doc-define` | `doc-define.yaml` | `doc-define-loop` ([doc-eval.md](doc-eval.md)) | `chains.docDefine` |
| `doc-eval` | `doc-eval.yaml` | `doc-eval-loop`, which also runs `doc-define` when a page has no definition ([doc-eval.md](doc-eval.md)) | `chains.docEval` |

The model picked in the chains tab comes first; see [Screen and API](#screen-and-api) and [config.md](config.md#goapp-config-file).

## Chain file

```yaml
name: doc-define
description: 제목과 본문 앞부분으로 문서가 다루는 대상을 한 줄로 정의한다.
version: v2.6
inputs:                         # filled by the task
  - title
  - wikitext
steps:
  - name: define
    message:
      system: |-
        당신은 지식 위키 '제타위키'의 문서 분류 담당자이다.
      user: |-
        ## 입력

        <bodyHead>
        {{ wikitext[:1000] }}
        </bodyHead>

        ## 과제
        제목이 '{{ title }}'인 문서의 본문 앞부분(`<bodyHead>`)을 보고, 이 문서가 다루는 대상이 무엇인지 짧게 정의한다.
        ...
    output_variable:
      format: json
```

Keys follow Semantic Kernel (snake_case); variable names are camelCase and start with a letter (`loop` is taken). The files are read at start with unknown fields rejected, and a missing file or step (`Chain.Step`) panics there.

### Variables

An input is a string the task fills, written as its name alone or with a `description`. A step's output is a string too.

### Steps

A step has a `name`, an optional `description` and its `message`; `output_variable` gives the `format` its answer is read in (`text` when omitted, or `json`) and a `name` when a later step reads it (doc-eval's `rubrics`). A step does not list its inputs: the variables its messages read are found when the chain loads, as LangChain finds a template's `input_variables`.

The doc-define template reads `{{ wikitext[:1000] }}`; doc-eval's judge reads the full wikitext, and the service checks quoted errors against that same text. Long documents therefore use more model context and tokens. Go keeps what the services need beyond the file: the structs a response is parsed into, values that check results (`DocEvalSafetyPolicies`), and the code between steps.

### Messages

A step's `message` has a `system` and a `user` key, each a jinja2 template written as a YAML block (`|-`). Either may be left out, but not both; they are sent system first. Each is parsed on its own (`prompts.ParseChat`), so a value can never move text into the other role, and an error names the key (`message.user: line 3, column 1: …`). A message that renders blank is not sent. Two messages of one role, or an assistant message, cannot be written; no chain needs them. Built-in steps name the page in the task sentence (`제목이 '{{ title }}'인 문서…`) and put the data they read at the top of the user message, in tags under `## 입력` (`<bodyHead>…</bodyHead>`). The title is one line and is what the task is about, so it reads best in the sentence; `제목이 '…'인` also needs no particle chosen by the title's last letter. Each built-in step puts the task and output instructions after its inputs in the user message; its system message supplies the role and context.

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

- Variable and output names are distinct.
- Every step's messages are jinja2, at least one of `system` and `user` is written, and they read only the chain's variables and the outputs of earlier steps. All of these are strings, so a `for` is refused.
- Every input is read by some step.
- `Format.Parse` strips a code fence around the whole response; `json` decodes into a struct, `text` trims. It checks the format only; the service checks what values mean.

## Running

`llmsvc.RunStep(ctx, model, chain, step, vars, out)` renders and runs; `RunMessages` runs messages rendered earlier. Either calls bob `/aigate` and parses; the chain names each call in the call history (`llm_calls`, see [goapp.md](goapp.md)). A failed parse is retried once; each attempt parses into a fresh value, `out` is set only on success, and the error keeps at most 500 characters of the response. When the retry call itself fails, its error keeps the parse error as `previous attempt`. The service code runs the steps in order and the code between them.

## Chains and tasks

A chain says what to ask a model: its templates, formats, outputs and version, and which values the task fills (inputs). It does not pick targets, read the wiki, or store results; its model is picked per chain (`services/chainsvc`). A goapp task decides when to run a chain and on what: it picks the target, fills the inputs, runs the steps through its service, stores the result, handles failures, and decides from the chain version whether a stored result can be reused. See the table above for which task runs which chain.

## Version

`version` is `v<major>.<minor>`, recorded in full in `llm_calls` and with each stored result, so calls and results can be told apart by minor. A stored result is current when its major matches (`prompts.SameMajor`; a version written before minors, such as `v2`, counts as `v2.x`):

- Minor (`v2.0 → v2.1`): most edits. Stored definitions, evaluations and rubrics stay current.
- Major (`v2.1 → v3.0`): a change to the shape of the inputs, format or output, or an edit whose earlier results should all be redone. Every stored result of the chain becomes stale.

The model is outside the chain and does not change the version. Any change to the inputs or steps needs a bump: `TestSpecVersions` (`goapp/prompts/builtin_test.go`) pins each chain's version to a hash of its inputs and steps, and fails until the version and its entry there are updated. It cannot tell a minor change from a major one; that is the author's call, and review's. See [doc-eval.md](doc-eval.md) for what doc-define and doc-eval reuse.

## Screen and API

The chains tab (`/tool/ai/chains`, sysop-only) is a table with a row per built-in chain, most recently used first: name, description, model, runs and last use. A row expands to the chain's version, inputs and steps (the variables each reads, its format and output, and its messages), with the model picker. `GET /api/ai/chains` returns each chain as in its file plus `model`, `uses` and `used_at` (the count and latest time of the successful calls of its last step in `llm_calls`).

Each chain runs on one model, picked in the chains tab (`services/chainsvc`, `app_settings` `chain_models`). Until a sysop picks one, the `GOAPP_CONFIG` file decides (`chains.<chain>.model`, then `chains.common.model`), then `config.DefaultModel`. Several chains may share a model; the models tab lists each model's chains.

- `GET /api/ai/chains`: the built-in chains, each with its `model`
- `PUT /api/ai/chains/{name}/model`: `{"model": "provider/model"}`; an empty model removes the saved one, so the file decides again
