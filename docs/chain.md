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

## Rules

- Inputs are chain inputs or earlier outputs; every placeholder is a declared input and every input is used (`chain_test.go`).
- Variable names are lowercase with `_`. System messages are fixed; changing values go into user messages.
- `Prompt.Render` inserts values as they are (a `{title}` inside a wikitext is not expanded) and leaves undeclared braces alone.
- `Parser.Parse`: `json` strips a code fence and decodes into a struct; `text` trims. It checks the format only; the service checks what values mean.

## Running

`llmsvc.RunStep(ctx, model, step, vars, out)` renders, calls bob `/aigate`, and parses. A failed parse is retried once; each attempt parses into a fresh value, `out` is set only on success, and the error keeps at most 500 characters of the response. When the retry call itself fails, its error keeps the parse error as `previous attempt`. The service code runs the steps in order and the code between them.

## Metadata

The scheme gives no key a meaning; each chain picks its own. doc-eval uses `version`, recorded in `doc_evaluations.chain_version` and bumped when a step's prompt, parser, or output changes (not the model). See [doc-eval.md](doc-eval.md).

## Screen and API

The chains tab (`/tool/ai/chains`, sysop-only) shows the built-in chains read-only, next to the aiedit prompts in the database. `GET /api/ai/builtin-chains` returns `prompts.BuiltinChains()`.
