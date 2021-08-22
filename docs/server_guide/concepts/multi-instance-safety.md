# Multi-instance safety

In its default configuration, the Sqedule server does not support running multiple concurrent instances. The main reason for this is because it [automatically migrates the database schema during startup](database-schema-migration.md), which is not concurrency-safe.

To make it safe to run multiple concurrent instances of the Sqedule server, [disable automatic database schema migration](../tasks/disabling-automatic-schema-migration.md).
