# Installation requirements

## Database

The Sqedule server requires PostgreSQL >= 13. Other databases may be supported in the future depending on user demand.

The PostgreSQL instance must have [the citext extension](https://www.postgresql.org/docs/14/citext.html) installed.

### Schemas

You don't need to manually setup database schemas. The Sqedule server [takes care of that automatically during startup](../concepts/database-schema-migration.md).

### Permissions

When the Sqedule server migrates the database schema, it will enable the citext extension. Since PostgreSQL 13, the citext extension is by default considered a trusted extension, and so anybody can enable it. Everything should work by default.

On older PostgreSQL versions, enabling the extension may result in this error:

~~~
ERROR:  permission denied to create extension "citext"
HINT:  Must be superuser to create this extension.
~~~

You can do one of the following to make it work:

 * Let a superuser role pre-enable the citext extension on the Sqedule database.

    ~~~sql
    CREATE EXTENSION citext;
    ~~~

 * -OR-: The PostgreSQL user that the Sqedule server authenticates with, must have the superuser role.

## Operating system

In theory, the server can be run on all operating systems that the Go programming language supports. But we've only tested on macOS and Linux, and we only provide precompiled binaries for Linux (x86\_64).
