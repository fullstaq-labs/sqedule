# Development guide

## Requirements

For developing the API backend, you need:

 * Go 1.15
 * PostgreSQL >= 13

For developing the web UI, you need:

 * Node.js >= 12 LTS

## Web UI one-time setup

Run:

~~~bash
cd webui
npm install
~~~

## Database setup

Create a PostgreSQL database, for example `sqedule_dev`. Then create the necessary database tables:

~~~bash
go run ./cmd/sqedule-server db migrate --db-type postgresql --db-connection 'dbname=sqedule_dev'
~~~

The argument passed to `--db-connection` is a string describing how to connect to your database. Run `go run ./cmd/sqedule-server db help` to learn more about the format.

### Seed data

The file `devtools/db-seed.sql` contains seed data for development purposes. This data includes organizations, user accounts, application and deployment definitions, etc.

You should load this seed data after having created the database tables. For example:

~~~bash
psql -d sqedule_dev -f devtools/db-seed.sql
~~~

### Database reset

The database migration command is idempotent, and only performs migrations that haven't already been run. If you wish to reset your database schema and data from scratch, then pass `--reset` to the database migration command, like this:

~~~bash
go run ./cmd/sqedule-server db migrate --db-type postgresql --db-connection 'dbname=sqedule_dev' --reset
~~~

You should reload the seed data after having done such a reset.

## Running the API server

~~~bash
go run ./cmd/sqedule-server server --db-type postgresql --db-connection dbname=sqedule_dev
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

First, authenticate as the `admin_sa` user account (who's part of the `org1` organization) and obtain an authentication token:

~~~bash
http POST localhost:3001/v1/auth/login service_account_name=admin_sa access_token=123456 organization_id=org1
~~~

This prints something like:

~~~
HTTP/1.1 200 OK
Content-Length: 253
Content-Type: application/json; charset=utf-8
Date: Tue, 05 Jan 2021 11:56:21 GMT

{
    "code": 200,
    "expire": "2021-02-17T04:56:21+01:00",
    "token": "<SOME TOKEN>"
}
~~~

Copy the token and put it in a shell variable, like this:

~~~bash
AUTH_TOKEN='<SOME TOKEN>'
~~~

Now you can request a list of releases:

~~~bash
http localhost:3001/v1/releases Authorization:"Bearer $AUTH_TOKEN"
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
