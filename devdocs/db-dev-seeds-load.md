# Loading database development seed data

The file `devtools/db-seeds.sql` contains useful development seed data. The data includes organizations, organization members, applications, rulesets, releases, and related bindings. You should load this data after having [initially set up the database](dev-env-setup.md), or after having [reset](db-reset.md) it.

To load it, run:

~~~bash
psql -d sqedule_dev -f devtools/db-seeds.psql
~~~

Rename "sqedule_dev" with the actual database name.

## See also

 * [Development seed data organization members](db-dev-seeds-org-members.md)
