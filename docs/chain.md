# Chains

goapp defines LLM work as chains, following LangChain. Types: `goapp/prompts/chain.go`; built-in chains are YAML files in `goapp/prompts/chains/`, one per chain (for example `doc-eval.yaml`), embedded into the binary.

| Term | Example | Meaning | LangChain |
| :--- | :--- | :--- | :--- |
| Chain | `doc-eval` | Steps run in order; a step's output feeds later steps | `RunnableSequence` |
| Step | `doc-eval/rubric` | One LLM call: prompt, model, parser, named output | `prompt \| model \| parser` (three `steps` there, one here) |
| Prompt | | Messages (`system`, `user`) with `{input}` placeholders | `ChatPromptTemplate` |
| Parser | `json`, `text` | Response text → value | output parser |
| Version | `v2` | Recorded with each call and result; the services decide from it whether a stored result is current | |

"System prompt" here is the system message; "user prompt" is not used.

## Built-in chains

| Chain | File | Run by | Model in the file |
| :--- | :--- | :--- | :--- |
| `doc-define` | `doc-define.yaml` | `doc-define-loop` ([doc-eval.md](doc-eval.md)) | `chains.docDefine` |
| `doc-eval` | `doc-eval.yaml` | `doc-eval-loop`, which also runs `doc-define` when a page has no definition ([doc-eval.md](doc-eval.md)) | `chains.docEval` |
| `doc-create` | `doc-create.yaml` | `ai-edit`, request type `create` | `chains.docCreate` |
| `doc-update` | `doc-update.yaml` | `ai-edit`, request type `update` | `chains.docUpdate` |

The model picked in the chains tab comes first; see [Screen and API](#screen-and-api) and [config.md](config.md#goapp-config-file). `doc-create` and `doc-update` pass through the prompt rendered from an AI edit prompt template, since the templates still live in `aiedit_prompts`.

## Chain file

```yaml
name: doc-eval
metadata:
  version: v2
  description: 제목과 대상 정의로 루브릭을 만들고, 그 루브릭과 본문으로 판정한다.
spec:
  inputs: [title, definition, wikitext]
  steps:
    - name: rubric
      description: ...
      inputs: [title, definition]
      messages:
        - role: system
          content: |-
            ...
        - role: user
          content: |-
            ### 문서 제목: {title}
            ### 대상 정의: {definition}
      parser: json
      output: rubrics
```

`metadata` describes the chain; `spec` is what decides its results. The files are read at start with unknown fields rejected, and a missing file or step (`Chain.Step`) panics there. Go keeps what the services need beyond the file: the structs a response is parsed into, values that check results (`DocEvalSafetyPolicies`), and the code between steps.

## Chains and tasks

A chain says what to ask a model: its prompts, parsers, outputs and version. It does not pick targets, read the wiki, or store results; its model is picked per chain (`services/chainsvc`). A goapp task decides when to run a chain and on what: it picks the target, gathers the inputs, runs the steps through its service, stores the result, handles failures, and decides from the chain version whether a stored result can be reused. See the table above for which task runs which chain.

## Rules

- Inputs are chain inputs or earlier outputs; every placeholder is a declared input and every input is used (`chain_test.go`).
- Variable names are lowercase with `_`. System messages are fixed; changing values go into user messages.
- `Prompt.Render` inserts values as they are (a `{title}` inside a wikitext is not expanded) and leaves undeclared braces alone.
- `Parser.Parse`: `json` strips a code fence and decodes into a struct; `text` trims. It checks the format only; the service checks what values mean.

## Running

`llmsvc.RunStep(ctx, model, chain, step, vars, out)` renders, calls bob `/aigate`, and parses; the chain names each call in the call history (`llm_calls`, see [goapp.md](goapp.md)). A failed parse is retried once; each attempt parses into a fresh value, `out` is set only on success, and the error keeps at most 500 characters of the response. When the retry call itself fails, its error keeps the parse error as `previous attempt`. The service code runs the steps in order and the code between them.

## Version

`metadata.version` is the chain's version, recorded in `llm_calls` and with each stored result. doc-define and doc-eval bump it when a step's prompt, parser, or output changes (not the model); see [doc-eval.md](doc-eval.md).

## Screen and API

The chains tab (`/tool/ai/chains`, sysop-only) is a table of the built-in chains (name and version, description, number of steps, model); `GET /api/ai/chains` returns each chain as in its file (`name`, `metadata`, `spec`) plus `model`. A row expands to its steps (inputs, parser, output, and messages, read-only), and for `doc-create` and `doc-update` to the AI edit prompt templates of that request type instead of their pass-through message. `/tool/ai/prompts` redirects there; a template's own pages stay under it.

Each chain runs on one model, picked in the chains tab (`services/chainsvc`, `app_settings` `chain_models`). Until a sysop picks one, the `GOAPP_CONFIG` file decides (`chains.<chain>.model`, then `chains.common.model`), then `config.DefaultModel`. Several chains may share a model; the models tab lists each model's chains.

- `GET /api/ai/chains`: the built-in chains, each with its `model`
- `PUT /api/ai/chains/{name}/model`: `{"model": "provider/model"}`; an empty model removes the saved one, so the file decides again
