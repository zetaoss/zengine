# Disambiguation Feature Guide

The disambiguation feature groups main-namespace pages that share a base title under one page in the `Disambig` namespace, then displays a list of alternatives at the top of each main-namespace page. Pages in other namespaces, including talk pages, are not treated as disambiguation targets. The disambiguation page is the source of truth. ZetaExtension maintains the display cache and page-relationship index, and ZetaSkin renders them.

## Page model

ZetaExtension registers these namespaces:

| Namespace | ID | Purpose |
| --- | ---: | --- |
| `Disambig` | 3002 | Disambiguation pages |
| `Disambig_talk` | 3003 | Disambiguation talk pages |

If a main-namespace page title ends in parentheses, its base title is the part before the final parenthetical suffix.

```text
사과       -> Disambig:사과
사과 (과일) -> Disambig:사과
사과(기업)  -> Disambig:사과
```

Each item in the first bulleted list of a disambiguation page uses this format:

```wikitext
* [[사과 (과일)|사과 (과일)]] - 과일
* [[사과 (기업)]] - 기업
```

The first direct link in each list item is the target page. The link text is its display name, and the text after ` - ` is its description. The UI editor only reads and writes this format, so it refuses table editing if the disambiguation page contains other wikitext.

## User flow

When an existing main-namespace page is viewed, ZetaExtension provides information about its base title and corresponding `Disambig` page through `disambigRegistration`.

- If no relationship exists yet, the page menu can open a modal to create or register a disambiguation page.
- When creating a new disambiguation page, the feature automatically finds main-namespace pages with the same base title and targets of existing `{{다른 뜻}}` templates.
- Existing disambiguation pages are loaded into a table that supports adding, deleting, reordering, and editing descriptions.
- Saving uses the MediaWiki edit API. Edits to existing pages include `basetimestamp` to detect conflicts.
- Main-namespace pages with a disambiguation relationship show the list and an edit button at the top of the page.
- Deleting a disambiguation page from the modal also removes its relationships through the page-delete hook.

An item title must either equal the base title or append a parenthetical suffix to it. For example, if the base title is `사과`, `사과`, `사과 (과일)`, and `사과(기업)` are allowed, but `풋사과` is not.

## Components

| Role | Path |
| --- | --- |
| Namespace, hooks, and job registration | `mwz/extensions/ZetaExtension/extension.json` |
| Base-title and registration data | `mwz/extensions/ZetaExtension/includes/Disambig/DisambigHooks.php` |
| Parsing, caching, and relationship synchronization | `mwz/extensions/ZetaExtension/includes/Disambig/DisambigService.php` |
| Save, delete, undelete, and move hooks | `mwz/extensions/ZetaExtension/includes/Collection/CollectionHooks.php` |
| Legacy template and base-title cleanup | `mwz/extensions/ZetaExtension/includes/Disambig/CleanupLegacyDisambigJob.php` |
| Full cache rebuild | `mwz/extensions/ZetaExtension/maintenance/RebuildDisambigs.php` |
| Page data lookup | `mwz/skins/ZetaSkin/includes/PageDataProvider.php` |
| Top-of-page list | `mwz/skins/ZetaSkin/svelte/src/components/disambig/DisambigApex.svelte` |
| Create and edit modal | `mwz/skins/ZetaSkin/svelte/src/components/disambig/DisambigModal.svelte` |

## Storage

The tables are currently in the `ldb` database.

### `disambigs`

Stores the display cache for each disambiguation page.

| Column | Meaning |
| --- | --- |
| `id` | MediaWiki page ID of the `Disambig` page |
| `cache` | JSON containing `id`, `text`, and `nodes` |
| `entries` | Number of items in the cache |
| `created_at` | Initial creation time |
| `updated_at` | Last synchronization time |

Each cache node contains a title, link text, URL, description, and either the `id` of an existing page or `new` to indicate a page that does not exist yet.

### `disambig_pages`

Stores relationships between main-namespace page titles and disambiguation pages.

| Column | Meaning |
| --- | --- |
| `disambig_id` | ID of the `Disambig` page that owns the relationship |
| `page_title` | Normalized database title of the target main-namespace page |
| `page_id` | Page ID when the target exists; `NULL` for a page that has not been created |

Titles are indexed in advance, so a relationship can be recalculated if a main-namespace page with the same title is created later, even when the disambiguation item does not yet exist. The current schema uses `page_title` as the primary key and keeps `page_id` unique, so a main-namespace page can belong to only one disambiguation page.

Creation SQL for both tables and compatibility handling that converts the existing development schema to a main-namespace-only structure are maintained together in `goapp/app/database/migrations/202608010001_collection_page_indexes.up.sql`. If the tables move to MediaWiki's primary database or become managed by an extension schema update, also update the `ldb.` references in the service and skin, and the component responsible for running the migration during deployment.

## One-time follow-up for migration `202608010001_collection_page_indexes`

Apply the migration with `tool` from the new application image:

```bash
tool migrate
```

After the initial migration, reparse all existing Binder pages to populate historical redlinks and redirect dependencies:

```bash
MW_INSTALL_PATH=/app/w php /app/mwz/extensions/ZetaExtension/maintenance/RebuildBinders.php --server localhost
```

The old PHP code is incompatible with the changed `binder_pages` schema. Suspend MediaWiki write requests and related job processing while applying the migration and switching the source code.

## Synchronization flow

`CollectionHooks` connects the MediaWiki page lifecycle to `DisambigService`.

```text
Save a Disambig page
  -> Parse the first list
  -> Replace disambig_pages relationships
  -> Upsert the disambigs cache
  -> Enqueue a legacy-cleanup job

Save or restore a main-namespace page
  -> Find relationships for the same not-yet-created title
  -> Rebuild the related Disambig cache

Move a page
  -> Recalculate relationships for the old page ID
  -> Recalculate relationships for the new title and any created redirect

Delete a page
  -> If it is a Disambig page, delete its cache and all relationships
  -> If it is a main-namespace page, rebuild the related cache
```

Relationships and cache entries are updated in one atomic section. The cache is derived data; if its contents look wrong, resave the disambiguation page or run the full rebuild command.

## Legacy disambiguation cleanup

Saving a disambiguation page registers a `zetaCleanupLegacyDisambig` job for each related main-namespace page.

- Removes line-level `{{다른 뜻}}` or `{{다른_뜻}}` templates from the main-namespace page body.
- If the base-title page is not in the list and is missing or a redirect, creates or updates a redirect to the first disambiguation item.
- Does not overwrite a base-title page that already has regular content.
- Edits as the user who saved the disambiguation page and records the edit as minor.
- Uses job-queue options to remove duplicate jobs for the same target.

Check job status and failures in the MediaWiki job queue. The job returns a failure if it cannot restore the editing user's identity or save the revision.

## Batch conversion of legacy disambiguation templates

Scans the wiki for pages that use the existing `{{다른뜻}}` / `{{다른 뜻}}` templates and processes them in batches.

```bash
# Dry run (preview the changes)
MW_INSTALL_PATH=/app/w php /app/mwz/extensions/ZetaExtension/maintenance/BatchConvertLegacyDisambigs.php --dry-run --server localhost

# Run the batch conversion
MW_INSTALL_PATH=/app/w php /app/mwz/extensions/ZetaExtension/maintenance/BatchConvertLegacyDisambigs.php --server localhost
```

### Conversion rules

1. Group pages containing `{{다른뜻}}` by base title.
2. Find existing main-namespace pages with each base title, excluding redirects.
3. **If at least two candidate pages exist, or a `Disambig` page already exists:**
   - Create `Disambig:<base title>` if it does not exist, then synchronize its database relationships and cache.
   - Remove the `{{다른뜻}}` template from the target page bodies.
4. **If fewer than two candidate pages exist:**
   - Do not create a `Disambig` page; only remove the `{{다른뜻}}` template from the target page bodies.

Options:

- `--dry-run`: List pages to create or change without making edits.
- `--user=<UserName>`: Set the username recorded as the migration editor. Defaults to `Jmnote bot`.
- `--limit=<N>`: Limit the number of base titles to process.

## Full rebuild

Reparses every `Disambig` page to rebuild its relationships and cache.

```bash
MW_INSTALL_PATH=/app/w php /app/mwz/extensions/ZetaExtension/maintenance/RebuildDisambigs.php
```

The command prints one line per successfully processed page and reports the total number at the end. If any page fails, it exits with a non-zero status after listing the failed page IDs, titles, and errors.

Consider a full rebuild:

- After creating or restoring the tables.
- After changing the parser or cache format.
- After disambiguation pages changed while the hooks were disabled.
- If `disambig_pages.page_id` may not match the actual MediaWiki page ID.

## Troubleshooting

If the disambiguation list does not appear, check the following in order:

1. Confirm that the corresponding `Disambig:<base title>` page exists and uses the supported bulleted-list format.
2. Confirm that `disambigs` contains valid JSON cache data for that Disambig page ID.
3. Confirm that `disambig_pages` contains a relationship for the current page's `page_id` or normalized title.
4. Check the `PageSaveComplete`, `PageDeleteComplete`, `PageUndeleteComplete`, and `PageMoveComplete` hooks and registration of the `zetaCleanupLegacyDisambig` job.
5. If needed, run the full rebuild command and inspect any failed pages.
