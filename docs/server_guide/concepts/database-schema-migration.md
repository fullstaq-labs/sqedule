# Database schema migration

During startup, the Sqedule server checks whether the database schema is up-to-date. If not then it will automatically migrate the database schema to the latest version.

Schema database migration is not concurrency-safe. This is the main reason why the Sqedule server currently [does not support running multiple concurrent instances](multi-instance-safety.md) in its default configuration.

You can [disable automatic database schema migration](../tasks/disabling-automatic-schema-migration.md). If you do this, then it is safe to run multiple concurrent Sqedule server instances, but you become responsible for [running database schema migrations manually](../tasks/manual-database-schema-migration.md) every time you upgrade the Sqedule server.
