# Running the server

## Prerequisites

Before you can run the server, you must [setup the development environment](dev-env-setup.md).

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
