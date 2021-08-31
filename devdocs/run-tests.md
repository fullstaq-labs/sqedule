# Running tests

## Prerequisites

Before you can run tests, you must [setup the development environment](dev-env-setup.md).

## Running the Go test suite

Set the following environment variables to tell the test suite where your [test database](dev-env-setup.md) is:

~~~bash
export SQEDULE_DB_TYPE=postgresql
export SQEDULE_DB_CONNECTION="dbname=sqedule_test"
~~~

Then run:

~~~bash
go test ./...
~~~
