# Manually migrating the database schema

By default, the Sqedule server [automatically migrates the database schema](../concepts/database-schema-migration.md) during startup. Sometimes you may want to migrate the schema manually:

 * You've [disabled automatic database schema migration](disabling-automatic-schema-migration.md).
 * You want to downgrade Sqedule to a previous version, and thus you need to rollback to a previous schema version.

## Upgrading the database schema

[Invoke the subcommand](../concepts/server-exe.md) `sqedule-server db migrate`. This subcommand requires the following [configuration options](../config/index.md):

 * `db-type`
 * `db-connection`

Example:

~~~bash
sqedule-server db migrate --db-type postgresql --db-connection 'dbname=sqedule user=sqedule password=something host=localhost port=5432'
~~~

## Rolling back the database schema

[Invoke the subcommand](../concepts/server-exe.md) `sqedule-server db rollback`. This subcommand requires the following [configuration options](../config/index.md).

 * `db-type`
 * `db-connection`
 * `target` (string) — The schema version to rollback to. This is something like "20210304000010 Release approval ruleset binding".

Example:

~~~bash
sqedule-server db rollback \
  --db-type postgresql \
  --db-connection 'dbname=sqedule user=sqedule password=something host=localhost port=5432' \
  --target '20210304000010 Release approval ruleset binding'
~~~

### Available schema target names

You can find out the available schema target names by [inspecting the Sqedule server's source code](https://github.com/fullstaq-labs/sqedule/tree/main/server/dbmigrations). Look in the subdirectory `server/dbmigrations` and you'll see files like these:

~~~
20201021000000_basic_settings.go
...
20210304000010_release_approval_ruleset_binding.go
~~~

Each one of these files is a schema target.

To find out the name of a specific target, look inside the file and look for `ID: "something"`. For example in `20210304000010_release_approval_ruleset_binding.go`:

~~~go
var migration20210304000010 = gormigrate.Migration{
	ID: "20210304000010 Release approval ruleset binding",
    ...
}
~~~

Here we see that the target name is "20210304000010 Release approval ruleset binding".

### Determining the schema target name for a previous Sqedule release

 1. [Visit Sqedule's Github repository](https://github.com/fullstaq-labs/sqedule/tree/main/server/dbmigrations).
 2. Click on Github's branch/tag dropdown, and select the tag for the Sqedule server version that you want to rollback the schema to.
 3. Browse to the subdirectory `server/dbmigrations`.
 4. Look inside the last migration file in the directory. The ID inside that file is the schema target name you're looking for.
