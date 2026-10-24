# Chains

goapp defines LLM work as chains, following LangChain. Types: `goapp/prompts/chain.go`; one file per chain (for example `goapp/prompts/doceval.go`).

| Term | Example | Meaning | LangChain |
| :--- | :--- | :--- | :--- |
| Chain | `doc-eval` | Steps run in order; a step's output feeds later steps | `RunnableSequence` |
| Step | `doc-eval/rubric` | One LLM call: prompt, model, parser, named output | `prompt \| model \| parser` (three `steps` there, one here) |
| Prompt | | Messages (`system`, `user`) with `{input}` placeholders | `ChatPromptTemplate` |
| Parser | `json`, `text` | Response text → value | output parser |
| Metadata | `version: v1` | Free key-value pairs on a chain | `RunnableConfig.metadata` |

"System prompt" here is the system message; "user prompt" is not used.

## Built-in chains

| Chain | File | Runs in | Model in the file |
| :--- | :--- | :--- | :--- |
| `doc-define` | `goapp/prompts/docdefine.go` | `doc-define-loop` ([doc-eval.md](doc-eval.md)) | `chains.docDefine` |
| `doc-eval` | `goapp/prompts/doceval.go` | `doc-eval-loop` ([doc-eval.md](doc-eval.md)) | `chains.docEval` |
| `doc-create` | `goapp/prompts/docwrite.go` | `ai-edit`, request type `create` | `chains.docCreate` |
| `doc-update` | `goapp/prompts/docwrite.go` | `ai-edit`, request type `update` | `chains.docUpdate` |

The model picked in the chains tab comes first; see [Screen and API](#screen-and-api) and [config.md](config.md#goapp-config-file). `doc-create` and `doc-update` pass through the prompt rendered from an AI edit prompt template, since the templates still live in `aiedit_prompts`.

## Rules

- Inputs are chain inputs or earlier outputs; every placeholder is a declared input and every input is used (`chain_test.go`).
- Variable names are lowercase with `_`. System messages are fixed; changing values go into user messages.
- `Prompt.Render` inserts values as they are (a `{title}` inside a wikitext is not expanded) and leaves undeclared braces alone.
- `Parser.Parse`: `json` strips a code fence and decodes into a struct; `text` trims. It checks the format only; the service checks what values mean.

## Running

`llmsvc.RunStep(ctx, model, chain, step, vars, out)` renders, calls bob `/aigate`, and parses; the chain names each call in the call history (`llm_calls`, see [goapp.md](goapp.md)). A failed parse is retried once; each attempt parses into a fresh value, `out` is set only on success, and the error keeps at most 500 characters of the response. When the retry call itself fails, its error keeps the parse error as `previous attempt`. The service code runs the steps in order and the code between them.

## Metadata

The scheme gives no key a meaning; each chain picks its own. doc-eval uses `version`, recorded in `doc_evaluations.chain_version` and bumped when a step's prompt, parser, or output changes (not the model). See [doc-eval.md](doc-eval.md).

## Screen and API

The chains tab (`/tool/ai/chains`, sysop-only) is a table of the built-in chains (name and version, description, number of steps, model). A row expands to its steps (inputs, parser, output, and messages, read-only), and for `doc-create` and `doc-update` to the AI edit prompt templates of that request type instead of their pass-through message. `/tool/ai/prompts` redirects there; a template's own pages stay under it.

Each chain runs on one model, picked in the chains tab (`services/chainsvc`, `app_settings` `chain_models`). Until a sysop picks one, the `GOAPP_CONFIG` file decides (`chains.<chain>.model`, then `chains.common.model`), then `config.DefaultModel`. Several chains may share a model; the models tab lists each model's chains.

- `GET /api/ai/chains`: the built-in chains, each with its `model`
- `PUT /api/ai/chains/{name}/model`: `{"model": "provider/model"}`; an empty model removes the saved one, so the file decides again
