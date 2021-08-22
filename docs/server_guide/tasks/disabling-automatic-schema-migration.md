# Disabling automatic schema migration

You may want to disable [automatic database schema migration](../concepts/database-schema-migration.md) in order to make the Sqedule server [multi-instance safe](../concepts/multi-instance-safety.md).

To disable, set the [configuration option](../config/index.md) `auto-db-migrate` to false.

!!! warning
    If you disable automatic database schema migration, then you become responsible for [running database schema migrations manually](manual-database-schema-migration.md) every time you upgrade the Sqedule server.
