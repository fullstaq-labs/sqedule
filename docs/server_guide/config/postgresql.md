# PostgreSQL connection string

PostgreSQL connection strings can be in DSN format or in URL format. Examples:

~~~yaml
db-connection: user=jack password=secret host=pg.example.com port=5432 dbname=mydb sslmode=verify-ca
db-connection: postgres://jack:secret@pg.example.com:5432/mydb?sslmode=verify-ca
~~~

Sqedule also reads from `PG*` environment variables, and (if a password is not provided) from the `.pgpass` file. Therefore, providing a connection string is optional.

Please refer to the PostgreSQL documentation on:

 - [Supported options in the connection string](https://www.postgresql.org/docs/13/libpq-connect.html#LIBPQ-PARAMKEYWORDS)
 - [The meaning of PG* environment variables](http://www.postgresql.org/docs/13/static/libpq-envars.html)

Note that Sqedule only supports these environment variables:

 - PGHOST
 - PGPORT
 - PGDATABASE
 - PGUSER
 - PGPASSWORD
 - PGPASSFILE
 - PGSERVICE
 - PGSERVICEFILE
 - PGSSLMODE
 - PGSSLCERT
 - PGSSLKEY
 - PGSSLROOTCERT
 - PGAPPNAME
 - PGCONNECT_TIMEOUT
 - PGTARGETSESSIONATTRS
