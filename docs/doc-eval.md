# Document Evaluation

An LLM evaluates ZetaWiki articles, treating an article as the response to its title. The structure follows Google's managed rubric-based metrics ([metrics](https://docs.cloud.google.com/gemini-enterprise-agent-platform/models/eval-python-sdk/determine-eval), [rubric details](https://docs.cloud.google.com/gemini-enterprise-agent-platform/models/rubric-metric-details), [templates](https://docs.cloud.google.com/gemini-enterprise-agent-platform/models/metrics-templates)); the Google service is not called.

Two chains ([chain.md](chain.md)) run in `goapp/services/doceval`. `doc-define` (`goapp/prompts/docdefine.go`) states what an article is about, so rubrics follow the article's subject rather than the title alone (a homonym or an abbreviation); `doc-eval` (`goapp/prompts/doceval.go`) grades it:

| Chain | Step | Inputs | Output |
| :--- | :--- | :--- | :--- |
| `doc-define` | `define` | `title`, `lead` (first 1500 bytes of the wikitext) | `definition`: one sentence on the article's subject |
| `doc-eval` | `rubric` | `title`, `definition` | `rubrics`: 5–8 general quality and 3–5 text quality criteria |
| `doc-eval` | `judge` | `title`, `rubrics`, `wikitext` (first 6000 bytes) | `judgment`: verdicts, safety, accuracy, summary |

## Items

| Item | Follows | Result |
| --- | --- | --- |
| General quality | `GENERAL_QUALITY` | Pass rate of the title's criteria (weight 80%) |
| Text quality | `TEXT_QUALITY` | Pass rate of coherence and fluency criteria (weight 20%) |
| Safety | `SAFETY` (safety_v3) + `spam`, `vandalism` | `safety_violated`, `safety_policies` |
| Accuracy | `GROUNDING`, adapted | Contradicted sentences |

- Criteria name aspects to cover, not facts, and never penalize length, list or table layout, recency, or missing citations. goapp assigns the ids (`g1`, `t1`, …); an omitted criterion fails.
- `vandalism` is deliberate damage; an honest mistake is an accuracy error. Personal data and copyright are not evaluated.
- Accuracy uses the judge's own knowledge (no `unsupported` label); opinions are `no_rad`. Only quotes found verbatim in the wikitext count.

## Grade

`F` on a safety violation, otherwise by composite score (80% general quality + 20% text quality): `S` ≥ 90%, `A` ≥ 75%, `B` ≥ 60%, `C` ≥ 40%, `D` below (`doceval.Grade`). Contradicted sentences are shown beside the grade and do not change it.

## Model and reuse

- Each chain has its model, picked in the evaluation tab (see [Settings](#settings)); bob does not fall back, so a quota-exhausted run fails and the next run retries.
- Each chain's metadata `version` is stored as `chain_version`. Rubrics are reused for the same title, definition and version; bump a version when its steps change, since results across versions are not comparable. A new version makes every article a candidate again.
- Redirect pages are skipped.

## Storage and API

- `doc_definitions` (`goapp/app/database/migrations/202610250001_doc_definitions.up.sql`): one row per article, `page_title`, `revid`, `definition`, model, `chain_version`. A definition is a property of the article, kept apart from evaluations so other features can use it; `doceval.EnsureDefinition` is its only writer, and it rewrites the row when the revision, title or chain version changes.
- `doc_evaluations` (`goapp/app/database/migrations/202610210001_doc_evaluations.up.sql`): grade, pass rates, safety, contradicted count, summary, the `definition` the rubrics were built from, `rubrics` and `raw_json` (judge output), `chain_version`, model, tokens, elapsed time.

All routes are sysop-only:
- `GET /api/evaluations`: list; filters `page_id`, `grade`, `contradicted=1`, `min_quality`, `max_quality`, `limit` (25, max 100), `offset`. Items include `revid`, `latest_revid`, `is_latest`, and `revisions_after` (count of edits made after the evaluation).
- `GET /api/evaluations/summary`: counts per grade, contradicted count, average pass rates
- `GET /api/evaluations/candidate`: picks a candidate article title by `strategy` (default: the setting)
- `GET /api/evaluations/{id}`: one evaluation, including revision status
- `GET`/`PUT /api/evaluations/settings`: `interval`, `model`, `strategy` (evaluation) and `define_interval`, `define_model`, `define_strategy` (definition); GET also returns `last_run`, `tasks`, `define_last_run`, `define_tasks`, and the number of `definitions`
- `POST /api/evaluations/run`: enqueues one evaluation run immediately (forces execution even when the interval is off)

## Loops

Both loops pick one article per run by their strategy (`random` by `page_random`, `oldest`/`newest` by `page_id`) and do nothing when no article is left:
- `doc-define-loop` defines an article without a current definition (none, or one of another revision, title or chain version).
- `doc-eval-loop` evaluates an article with a current definition and no evaluation of its latest revision by the current chain version. An edit makes the article a candidate again once it is redefined. A run for a given title or page ID (`POST /api/evaluations/run`) writes the definition first when it is missing.

Each run fetches the wikitext, runs its chain, saves the result and the run (`last_run`, `define_last_run`), and schedules the next one after its interval. Runs are never retried: a failed run is archived, recorded, and the loop goes on.

The loops share their own `doc-eval` queue, and each keeps at most one waiting task. `goapp/app/loopq` makes every change to that task under a Redis lock shared by workers and servers:
- A finished run schedules the next one unless a run is already waiting, so extra waiting runs collapse into one. A run canceled by worker shutdown is requeued by asynq and schedules nothing itself.
- `POST /api/evaluations/run` and an interval change enqueue the new run first and then delete the other waiting runs; a failed enqueue leaves the loop as it was. Interval `0s` deletes the waiting run.
- The worker on start and the hourly `system-loop-nanny` enqueue a run when none is waiting or active.

## Settings

Each field resolves in this order: the value a sysop saved in the evaluation tab (`app_settings` `doc_eval_settings`), `chains.docEval` or `chains.docDefine` of the `GOAPP_CONFIG` file, `chains.common`, then the built-in default. A value that does not parse falls back to the next.

| Field | Built-in default | Values |
| --- | --- | --- |
| `interval` | `1h` (docEval), `30m` (docDefine) | Go duration below 100h; `0s` turns the loop off |
| `model` | `groq/openai/gpt-oss-120b` | A bob aigate model, `provider/model` |
| `strategy` | `random` | `random`, `oldest`, `newest` |

```yaml
chains:
  common:
    model: groq/openai/gpt-oss-120b
    strategy: random
  docDefine:
    interval: 30m
  docEval:
    interval: 1h
```
