# DB Migrations

This directory contains SQL migration files used to manage the database schema.

## Filename Convention

All migration files must follow this naming convention:

```
YYYYMMDD_NNN_description.up.sql
```

- **`YYYYMMDD`**: The creation date of the migration (Year, Month, Day).
- **`NNN`**: A 3-digit sequential number starting from `001` (e.g., `001`, `002`, `003`), reset daily. Use this to order migrations created on the same day.
- **`description`**: A concise, snake_case description of what the migration does (e.g., `add_pageid_to_aiedit_tasks`).
- **`.up.sql`**: The file suffix required for "up" (apply) migrations.

### Examples

- `20260620_001_add_pageid_to_aiedit_tasks.up.sql`
- `20260620_002_create_users_table.up.sql`

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

### Automatic Error Handling

MySQL does not support transactional DDL (e.g., `ALTER TABLE` commits implicitly). If a migration file containing both schema changes and data updates fails halfway (e.g., during the data update), re-running the migration would normally fail on the DDL statements.

To address this, **the migration runner is designed with built-in resilience** and will automatically ignore specific MySQL error codes when re-running a migration:

- **`1054` (Unknown column)**: Occurs when trying to modify or rename a column that has already been altered.
- **`1060` (Duplicate column name)**: Occurs when trying to add a column that already exists.
- **`1061` (Duplicate key name)**: Occurs when trying to add an index that already exists.
- **`1091` (Can't DROP column/key; check that it exists)**: Occurs when trying to drop a column or key that doesn't exist.

### Writing Best Practices

1. **Keep it as idempotent as possible**: Although the runner ignores common DDL duplicate/missing errors, always write your data manipulation queries (`UPDATE`, `INSERT`, etc.) defensively (e.g., checking if columns are empty or using `INSERT IGNORE` / `ON DUPLICATE KEY UPDATE` if applicable).
2. **Order of statements**: Place your schema changes (`ALTER TABLE`) before data updates. If the data update fails, the runner can safely skip the completed schema changes on the next run and retry the data update.

