# Document Evaluation

An LLM evaluates ZetaWiki articles, treating an article as the response to its title. The structure follows Google's managed rubric-based metrics ([metrics](https://docs.cloud.google.com/gemini-enterprise-agent-platform/models/eval-python-sdk/determine-eval), [rubric details](https://docs.cloud.google.com/gemini-enterprise-agent-platform/models/rubric-metric-details), [templates](https://docs.cloud.google.com/gemini-enterprise-agent-platform/models/metrics-templates)); the Google service is not called.

It is the `doc-eval` chain ([chain.md](chain.md)) in `goapp/prompts/doceval.go`, run by `goapp/services/doceval`:

| Step | Inputs | Output |
| :--- | :--- | :--- |
| `rubric` | `title` | `rubrics`: 5–8 general quality and 3–5 text quality criteria |
| `judge` | `title`, `rubrics`, `wikitext` (first 6000 bytes) | `judgment`: verdicts, safety, accuracy, summary |

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

- The model is picked in the evaluation tab (default `groq/openai/gpt-oss-120b`) and used for both steps; bob does not fall back, so a quota-exhausted run fails and the next run retries.
- The chain's metadata `version` is stored as `chain_version`. Rubrics are reused for the same title and version; bump the version when either step changes, since scores across versions are not comparable.
- Redirect pages are skipped.

## Storage and API

Table `doc_evaluations` (`goapp/app/database/migrations/202610210001_doc_evaluations.up.sql`): grade, pass rates, safety, contradicted count, summary, `rubrics` and `raw_json` (judge output), `chain_version`, model, tokens, elapsed time.

All routes are sysop-only:
- `GET /api/evaluations`: list; filters `page_id`, `grade`, `contradicted=1`, `min_quality`, `max_quality`, `limit` (25, max 100), `offset`. Items include `revid`, `latest_revid`, `is_latest`, and `revisions_after` (count of edits made after the evaluation).
- `GET /api/evaluations/summary`: counts per grade, contradicted count, average pass rates
- `GET /api/evaluations/candidate`: picks a random candidate article title (unreviewed or oldest reviewed)
- `GET /api/evaluations/{id}`: one evaluation, including revision status
- `GET`/`PUT /api/evaluations/settings`: `interval` (Go duration < 100h, `0s` = off, default `1h`) and `model` (default `groq/openai/gpt-oss-120b`); GET also returns `last_run` and queued loop tasks
- `POST /api/evaluations/run`: enqueues one evaluation run immediately (forces execution even when the interval is off)

## Loop

`doc-eval-loop` picks a non-redirect article (never-evaluated first), fetches its wikitext, runs the chain, saves the result, and schedules itself after the interval. It is never retried: a failed run is archived, recorded in `last_run`, and the loop goes on.

The loop runs in its own `doc-eval` queue and keeps at most one waiting task. `goapp/app/loopq` makes every change to that task under a Redis lock shared by workers and servers:
- A finished run schedules the next one unless a run is already waiting, so extra waiting runs collapse into one. A run canceled by worker shutdown is requeued by asynq and schedules nothing itself.
- `POST /api/evaluations/run` and an interval change enqueue the new run first and then delete the other waiting runs; a failed enqueue leaves the loop as it was. Interval `0s` deletes the waiting run.
- The worker on start and the hourly `system-loop-nanny` enqueue a run when none is waiting or active.
