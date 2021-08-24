# Development guide

## Requirements

For developing the API backend, you need:

 * Go 1.16
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

## Database setup

Create a PostgreSQL database, for example `sqedule_dev`. Then create the necessary database tables:

~~~bash
./devtools/server db migrate --db-type postgresql --db-connection 'dbname=sqedule_dev'
~~~

The argument passed to `--db-connection` is [a string describing how to connect to your database](https://docs.sqedule.io/server_guide/config/postgresql/).

### Seed data

The file `devtools/db-seed.sql` contains seed data for development purposes. This data includes organizations, user accounts, application and deployment definitions, etc.

You should load this seed data after having created the database tables. For example:

~~~bash
psql -d sqedule_dev -f devtools/db-seed.sql
~~~

### Database reset

The database migration command is idempotent, and only performs migrations that haven't already been run. If you wish to reset your database schema and data from scratch, then pass `--reset` to the database migration command, like this:

~~~bash
./devtools/server -tags dev db migrate --db-type postgresql --db-connection 'dbname=sqedule_dev' --reset
~~~

You should reload the seed data after having done such a reset.

## Running the API server

~~~bash
./devtools/server -tags dev run --db-type postgresql --db-connection dbname=sqedule_dev --dev
~~~

The API server listens on http://localhost:3001 by default.

## Running the web UI

~~~bash
cd webui
npm run dev
~~~

The web UI listens on http://localhost:3000 by default.

It assumes that the API server is listening on http://localhost:3001.

## Example: making an API request

This example shows how you can request a list of releases from the API server. We assume that:

 * You have [httpie](https://httpie.io/) installed.
 * You have the seed data loaded.

Request a list of releases:

~~~bash
http localhost:3001/v1/releases
~~~

~~~
HTTP/1.1 200 OK
Content-Length: 905
Content-Type: application/json; charset=utf-8
Date: Tue, 05 Jan 2021 11:59:39 GMT

{
    "items": [
        {
        	...
        }
    ]
}
~~~
