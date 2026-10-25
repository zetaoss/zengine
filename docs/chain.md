# Chains

goapp defines LLM work as chains, following LangChain. Types: `goapp/prompts/chain.go`; built-in chains are YAML files in `goapp/prompts/chains/`, one per chain (for example `doc-eval.yaml`), embedded into the binary.

| Term | Example | Meaning | LangChain |
| :--- | :--- | :--- | :--- |
| Chain | `doc-eval` | Steps run in order; a step's output feeds later steps | `RunnableSequence` |
| Step | `doc-eval/rubric` | One LLM call: prompt, model, format, and an output name when a later step reads it | `prompt \| model \| parser` (three `steps` there, one here); output name like ADK `output_key` |
| Prompt | | Messages (`system`, `user`) with `{name}` placeholders, each optionally sent only `if` a variable is given or once for `each` item of a list | `ChatPromptTemplate` (f-string); `if`/`each` like `MessagesPlaceholder` |
| User inputs | `hint`, `materials` | What the requester gives in the run form; they reach the messages as they are | graph input schema (LangGraph) |
| Inputs | `title`, `wikitext` | What the task running the chain fills | |
| Format | `json` (`text` when omitted) | What the step asks the model to return; the response is read in it | `response_format` (OpenAI), output parser |
| Version | `v2.0` | `v<major>.<minor>`, recorded with each call and result; a stored result of the same major is current | |

"System prompt" here is the system message; "user prompt" is not used.

## Built-in chains

| Chain | File | Run by | Model in the file |
| :--- | :--- | :--- | :--- |
| `doc-define` | `doc-define.yaml` | `doc-define-loop` ([doc-eval.md](doc-eval.md)) | `chains.docDefine` |
| `doc-eval` | `doc-eval.yaml` | `doc-eval-loop`, which also runs `doc-define` when a page has no definition ([doc-eval.md](doc-eval.md)) | `chains.docEval` |
| `doc-create` | `doc-create.yaml` | `ai-edit`, request type `create` | `chains.docCreate` |
| `doc-update` | `doc-update.yaml` | `ai-edit`, request type `update` | `chains.docUpdate` |

The model picked in the chains tab comes first; see [Screen and API](#screen-and-api) and [config.md](config.md#goapp-config-file). For `doc-create` and `doc-update` the requester may pick an AI edit prompt template (`aiedit_prompts`), which runs with its own run form and messages: `user_inputs` in place of the chain's user inputs (`Chain.WithUserInputs`: any names but the task's inputs and `index`; values reach the messages as they are; `readBacklinks` also has the task read the pages that link here) and `messages` in place of the step's (`Chain.ValidateMessages`: the rules of `Validate`, except that it need not use every input). A template saved before it had messages runs its content in place of the step's unconditional messages, followed by the step's `if` and `each` messages (`aiedit.TemplateMessages`).

## Chain file

```yaml
name: doc-create
metadata:
  version: v2.0
  description: 새 문서의 위키텍스트를 작성한다.
spec:
  userInputs:                 # the run form
    - name: hint
      label: 힌트
    - name: readBacklinks
      label: 역링크 읽기
      type: bool
    - name: materials
      label: 자료
      type: texts
  inputs:                     # filled by the task
    - name: title
    - name: backlinkTexts
      type: texts
  steps:
    - name: generate
      inputs: [title, hint, materials, backlinkTexts]
      messages:
        - role: system
          content: |-
            ...
        - role: user
          content: "### 문서 제목: {title}"
        - role: user
          if: hint
          content: "### 힌트: {hint}"
        - role: user
          each: materials
          content: |-
            ### 자료 {index}:
            {materials}
```

`metadata` describes the chain; `spec` is what decides its results. A variable's `type` is `string` (the default), `bool` or `texts` (a list of strings); an input may be written as its name alone (`inputs: [title, lead]`). A step names its `output` only when a later step reads it (doc-eval's `rubrics`), and writes `format: json` only when it takes JSON. The files are read at start with unknown fields rejected, and a missing file or step (`Chain.Step`) panics there. Go keeps what the services need beyond the file: the structs a response is parsed into, values that check results (`DocEvalSafetyPolicies`), and the code between steps.

## Chains and tasks

A chain says what to ask a model: its prompts, formats, outputs and version, and which values the requester gives (user inputs) and the task fills (inputs). It does not pick targets, read the wiki, or store results; its model is picked per chain (`services/chainsvc`). A goapp task decides when to run a chain and on what: it picks the target, fills the inputs, runs the steps through its service, stores the result, handles failures, and decides from the chain version whether a stored result can be reused. See the table above for which task runs which chain.

`ai-edit` fills `title`, `wikitext` (the page as it stands in the editor, else the saved page) and `backlinkTexts`: with `readBacklinks` on, up to 10 main-namespace articles linking to the page (the page may not exist yet), each as its title and the text around its first link to the page, references and comments removed, or its beginning when the link comes from a template (`services/wikisvc`).

## Rules

`Chain.Validate` checks these at start (`mustLoad`), and in `chain_test.go`:

- A step's inputs are variables or earlier outputs, and each is used by its messages as a placeholder, `if` or `each`. Every task input is used; a user input need not be (`readBacklinks` only tells the task to read).
- Placeholders, `if` and `each` name the step's inputs. `each` names a `texts` variable, and inside its message `{name}` is the item and `{index}` its number from 1; a `texts` variable is a placeholder only there.
- Variable names start with a letter (`backlinkTexts`, `wikitext`). System messages are fixed; changing values go into user messages.
- `Prompt.Render` skips a message whose `if` variable is not given (a blank string, false, an empty list), sends an `each` message per item, and inserts values as they are (a `{title}` inside a wikitext is not expanded); braces that are not variables stay: `{|`, JSON, and wiki templates, since a `{name}` right after another brace (`{{title}}`, `{{Infobox}}`, `{{{1}}}`) is not a placeholder, while one inside a template's parameter (`{{틀|{title}}}`) is.
- `Format.Parse` strips a code fence around the whole response; `json` decodes into a struct, `text` trims. It checks the format only; the service checks what values mean.

## Running

`llmsvc.RunStep(ctx, model, chain, step, vars, out)` renders and runs; `RunMessages` runs messages rendered earlier (AI edit stores them with the task). Either calls bob `/aigate` and parses; the chain names each call in the call history (`llm_calls`, see [goapp.md](goapp.md)). A failed parse is retried once; each attempt parses into a fresh value, `out` is set only on success, and the error keeps at most 500 characters of the response. When the retry call itself fails, its error keeps the parse error as `previous attempt`. The service code runs the steps in order and the code between them.

## Version

`metadata.version` is `v<major>.<minor>`, recorded in full in `llm_calls` and with each stored result, so calls and results can be told apart by minor. A stored result is current when its major matches (`prompts.SameMajor`; a version written before minors, such as `v2`, counts as `v2.x`):

- Minor (`v2.0 → v2.1`): most edits. Stored definitions, evaluations and rubrics stay current.
- Major (`v2.1 → v3.0`): a change to the spec's shape (inputs, format, output), or an edit whose earlier results should all be redone. Every stored result of the chain becomes stale.

The model is outside the spec and does not change the version. Any spec change needs a bump: `TestSpecVersions` (`goapp/prompts/builtin_test.go`) pins each chain's version to a hash of its spec and fails until the version and its entry there are updated. It cannot tell a minor change from a major one; that is the author's call, and review's. See [doc-eval.md](doc-eval.md) for what doc-define and doc-eval reuse.

## Screen and API

The chains tab (`/tool/ai/chains`, sysop-only) is a table with a row per built-in chain and per AI edit prompt template: owner (`-` for a built-in), name, description, model, runs and last use. `GET /api/ai/chains` returns each chain as in its file (`name`, `metadata`, `spec`) plus `model`, `uses` and `used_at` (the count and latest time of the successful calls of its last step in `llm_calls`).

- Built-in chains come first, on a gray background, most recently used first. A row expands to its version, run form (as it will look), inputs and steps (inputs, format, output, and messages with their `if` and `each`, read-only), with the model picker.
- Templates of `doc-create` and `doc-update` follow, most recently used first (`used_at`, recorded from this release on), then by use count; the requester's own are on light blue. A template runs on the model it picked (none yet: the chain's model). A row expands to its version and the messages it runs with, with the model picker (owner only), edit and delete (owner only), and clone.
- Editing turns the title and description into fields in place, and the run form and messages into one YAML field (`prompts.Template`: `userInputs` and `messages`, written like the chain files), read on save with unknown fields rejected and checked against the chain. Beside it, `POST /api/ai-prompts/check` shows what the server reads, shortly after typing stops: the run form as it will look, and the messages, or the error that would stop the save. Changing the run form or messages bumps the version (`v1.0` on creation; minor by default, major when picked); title, description and model do not. Each run records it as `prompt_version`.
- Clone opens the editor in a dialog, filled with the source's values and the title with ` (복제본)`, and creates the template on save. Cloning a built-in chain starts from its run form and first step's messages as they are (`template` in `GET /api/ai/chains`), with the chain's description and model.
- `/tool/ai/prompts` and the old template pages redirect to the tab.

Each chain runs on one model, picked in the chains tab (`services/chainsvc`, `app_settings` `chain_models`). Until a sysop picks one, the `GOAPP_CONFIG` file decides (`chains.<chain>.model`, then `chains.common.model`), then `config.DefaultModel`. Several chains may share a model; the models tab lists each model's chains.

- `GET /api/ai/chains`: the built-in chains, each with its `model`
- `PUT /api/ai/chains/{name}/model`: `{"model": "provider/model"}`; an empty model removes the saved one, so the file decides again
