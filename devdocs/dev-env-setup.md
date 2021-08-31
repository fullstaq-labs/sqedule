# Development environment setup

## Requirements

For developing the API backend, you need:

 * Go >= 1.16
 * PostgreSQL >= 13

For developing the web UI, you need:

 * Node.js >= 12 LTS
 * NPM >= 7

## Web UI one-time setup

Run:

~~~bash
cd webui
npm install
~~~

## Development database setup

Create a PostgreSQL database, for example `sqedule_dev`. Then create the necessary database tables:

~~~bash
./devtools/server db migrate --db-type postgresql --db-connection 'dbname=sqedule_dev'
~~~

The argument passed to `--db-connection` is [a string describing how to connect to your database](https://docs.sqedule.io/server_guide/config/postgresql/).

After setting up the database, you should [load it with development seed data](db-dev-seeds-load.md).

## Test database setup

Create another PostgreSQL database, for example `sqedule_test`. This database is used for [running tests](run-tests.md).

Then create the necessary database tables:

~~~bash
./devtools/server db migrate --db-type postgresql --db-connection 'dbname=sqedule_test'
~~~

Don't load [development seed data](db-dev-seeds-load.md) into the test database. It's pointless: the test suite deletes all data before every test.

## See also

 * [Running the server](running-the-server.md)
