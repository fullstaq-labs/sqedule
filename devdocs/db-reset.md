# Resetting the database

To reset the database, run the server's `db migrate` subcommand with the `--reset` flag. For example:

~~~bash
./devtools/server db migrate --db-type postgresql --db-connection 'dbname=sqedule_dev' --reset
~~~

With `--reset` enabled, `db migrate` drops all tables and recreates the schema from scratch.

After resetting the database, you should [load it with development seed data](db-dev-seeds-load.md).
