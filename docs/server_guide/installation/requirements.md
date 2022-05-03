# Installation requirements

## Database

The Sqedule server requires PostgreSQL >= 13. Other databases may be supported in the future depending on user demand.

The PostgreSQL instance must have [the citext extension](https://www.postgresql.org/docs/14/citext.html) installed.

### Schemas

You don't need to manually setup database schemas. The Sqedule server [takes care of that automatically during startup](../concepts/database-schema-migration.md).

### Permissions

When the Sqedule server migrates the database schema, it will enable the citext extension. This requires sufficient permissions. Therefore the Sqedule server requires one of the following:

 * The database must be owned by the PostgreSQL user that the Sqedule server authenticates with.

    !!! tip
        You can change the owner with:

        ~~~sql
        ALTER DATABASE your-database-name OWNER TO your-user-name;
        ~~~

 * -OR- (since PostgreSQL 13): The citext extension must be marked as a trusted extension.

## Operating system

In theory, the server can be run on all operating systems that the Go programming language supports. But we've only tested on macOS and Linux, and we only provide precompiled binaries for Linux (x86\_64).
