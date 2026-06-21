# DB Migrations

This directory contains SQL migration files used to manage the database schema.

## Filename Convention

All migration files must follow this naming convention:

```
YYYYMMDDNNNN_description.up.sql
```

- **`YYYYMMDD`**: The creation date of the migration (Year, Month, Day).
- **`NNNN`**: A 4-digit sequential number starting from `0001` (e.g., `0001`, `0002`, `0003`), reset daily. It is part of the version before the first underscore, so every migration has a unique version.
- **`description`**: A concise, snake_case description of what the migration does (e.g., `add_pageid_to_aiedit_tasks`).
- **`.up.sql`**: The file suffix required for "up" (apply) migrations.

### Examples

- `202606200001_add_pageid_to_aiedit_tasks.up.sql`
- `202606200002_create_users_table.up.sql`

## Execution

These migrations are packaged into the application binary using Go's `embed` package (defined in `goapp/app/database/migrate.go`).

To apply pending migrations, run the following command in your terminal within the application context:

```bash
ctl migrate
```

The runner will check the `schema_migrations` table and apply any unapplied migration files sequentially in order of their filename version.

## Idempotency & Error Handling

### The Role of Transactions

The migration runner wraps the execution of each migration file in a database transaction (`BEGIN`, `COMMIT`/`ROLLBACK`). However, it's critical to understand how this interacts with MySQL:

- **DDL statements (`ALTER TABLE`, `CREATE TABLE`, etc.) cause an implicit commit.** This means if you have DDL in your script, the transaction is committed at that point, and it cannot be rolled back.
- **Transactions remain crucial for DML (`INSERT`, `UPDATE`, etc.).** If a script contains multiple data updates and one fails, the transaction ensures all data changes within that script are rolled back, preventing partial data updates. It also ensures the final step—inserting the record into the `schema_migrations` table—is atomic with the data changes.

### Idempotent DDL

MySQL does not support transactional DDL (e.g., `ALTER TABLE` commits implicitly). If a migration file containing both schema changes and data updates fails halfway (e.g., during the data update), re-running the migration would normally fail on the DDL statements.

To address this, DDL statements must be idempotent. Use native `IF EXISTS`/`IF NOT EXISTS` clauses when supported. Otherwise, query `information_schema` and execute the DDL conditionally. The migration runner does not suppress database errors because the same MySQL error codes can also indicate failed data changes.

### Writing Best Practices

1. **Make DDL idempotent**: Guard schema changes with `IF EXISTS`/`IF NOT EXISTS` or `information_schema` checks so a partially applied migration can be retried safely.
2. **Write DML defensively**: Use precise predicates and, where appropriate, `INSERT IGNORE` or `ON DUPLICATE KEY UPDATE` so retries do not duplicate or overwrite data unexpectedly.
3. **Order of statements**: Place schema changes before data updates. If a data update fails, a retry can safely skip the completed idempotent schema changes and retry the data update.
