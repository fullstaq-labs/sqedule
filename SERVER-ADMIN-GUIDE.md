# Sqedule server administration guide

## Installation

### Requirements

The Sqedule server requires PostgreSQL >= 13. Other databases may be supported in the future depending on user demand.

You don't need to manually setup database schemas. The Sqedule server [takes care of that automatically during startup](#database-schema-migration).

### Installing the binary (without containerization)

 1. Create a user account to run the Sqedule server in. For example:

    ~~~basH
    sudo addgroup --gid 3420 sqedule-server
    sudo adduser --uid 3420 --gid 3420 --disabled-password --gecos 'Sqedule Server' sqedule-server
    ~~~

 2. [Download a Sqedule server binary tarball](https://github.com/fullstaq-labs/sqedule/releases) (`sqedule-server-XXX-linux-amd64.tar.gz`).

    Extract the tarball. There's a `sqedule-server` executable inside. Check whether it works:

    ~~~bash
    /path-to/sqedule-server --help
    ~~~

 3. Create a Sqedule server configuration file `/etc/sqedule-server.yml`. Learn more in [Configuration](#configuration).

    At minimum you need to configure the database type and credentials. Example:

    ~~~yaml
    db-type: postgresql
    db-connection: 'dbname=sqedule user=sqedule password=something host=localhost port=5432'
    ~~~

    Be sure to give the file the right permissions so that the database password cannot be read by others:

    ~~~bash
    sudo chown sqedule-server: /etc/sqedule-server.yml
    sudo chmod 600 /etc/sqedule-server.yml
    ~~~

 4. Install a SystemD service file. Create /etc/systemd/system/sqedule-server.service:

    ~~~systemd
    [Unit]
    Description=Sqedule Server

    [Service]
    ExecStart=/path-to/sqedule-server run --config=/etc/sqedule-server.yml
    User=sqedule-server
    PrivateTmp=true

    [Install]
    WantedBy=multi-user.target
    ~~~

    (Be sure to replace `/path-to/sqedule-server`!)

    Then:

    ~~~bash
    sudo systemctl daemon-reload
    ~~~

 5. Start the Sqedule server:

    ~~~bash
    sudo systemctl start
    ~~~

    > **Note**: you don't need to manually setup database schemas. The Sqedule server [takes care of that automatically during startup](#database-schema-migration).

    It listens on localhost port 3001 by default. Try it out, you should see the web interface HTML:

    ~~~bash
    curl localhost:3001
    ~~~

### Installing the container

Use the image `ghcr.io/fullstaq-labs/sqedule-server`.

 * You must pass configuration via environment variables. Learn more in [Configuration](#configuration). At minimum you need to configure the database type and credentials.
 * Inside the container, the Sqedule server listens on port 3001.
 * You don't need to manually setup database schemas. The Sqedule server takes care of that automatically during startup.

Example:

~~~bash
docker run --rm \
  -p 3001:3001 \
  -e SQEDULE_DB_TYPE=postgresql \
  -e SQEDULE_DB_CONNECTION='dbname=sqedule user=sqedule password=something host=localhost port=5432' \
  ghcr.io/fullstaq-labs/sqedule-server
~~~

Try it out, you should see the web interface HTML:

~~~bash
curl localhost:3001
~~~

### Installing with Kubernetes

 1. Create a secret containing the database location and credentials.

    First, create a database connection string and encode it as Base64:

    ~~~
    echo -n 'dbname=sqedule user=sqedule password=something host=localhost port=5432' | base64
    ~~~

    Then create a Kubernetes Secret:

    ~~~yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: sqedule-db-connection
    type: Opaque
    data:
      db_connection: <BASE64 DATA HERE>
    ~~~

 2. Create a Kubernetes Deployment.

     * You must pass configuration via environment variables. Learn more in [Configuration](#configuration). At minimum you need to configure the database type and credentials.
     * You don't need to manually setup database schemas. The Sqedule server takes care of that automatically during startup.
     * Sqedule in its default configuration does not support running multiple instances. Therefore, unless you've taken steps to [make Sqedule multi-instance-safe](#running-multiple-instances-of-sqedule), you must only run a single replica and you must only use the `recreate` update strategy.

    Example:

    ~~~yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: sqedule
      labels:
        app: sqedule
    spec:
      selector:
        matchLabels:
          app: sqedule
      strategy:
        type: recreate
      template:
        metadata:
          labels:
            app: sqedule
        spec:
          containers:
            - name: sqedule
              image: ghcr.io/fullstaq-labs/sqedule-server
              ports:
                - containerPort: 3001
              env:
                - name: SQEDULE_DB_TYPE
                  value: postgresql
                - name: SQEDULE_DB_CONNECTION
                  valueFrom:
                    secretKeyRef:
                      name: sqedule-db-connection
                      key: db_connection
    ~~~

 3. Create a Kubernetes Service so that the Sqedule server can be accessed through the network.

    ~~~yaml
    apiVersion: v1
    kind: Service
    metadata:
      name: sqedule
    spec:
      selector:
        app: sqedule
      ports:
        - protocol: TCP
          port: 80
          targetPort: 3001
    ~~~

## About the Sqedule server executable

The Sqedule server comes with an executable named `sqedule-server`. This executable supports multiple subcommands, similar to how Git works. The most often used subcommand is `run`, which runs the HTTP server.

You normally don't need to interact with this executable directly, because you're either using this executable through the SystemD service file, or through a container. This section explains how to interact with it anyway, in case you need to troubleshoot something or in case you want to understand how it works.

### Invoking the Sqedule server executable

The way to invoke the Sqedule server executable depends on the installation method.

 - When installed directly as a binary (without containerization), invoke `sqedule-server` directly from your shell. Example:

   ~~~bash
   sqedule-server --help
   ~~~

   If `sqedule-server` is not in PATH, then invoke its full path: `/path-to/sqedule-server`

 - When using containerization, invoke it through a container. The image's entrypoint is the Sqedule server executable. Example:

   ~~~bash
   docker run -ti --rm ghcr.io/fullstaq-labs/sqedule-server --help
   ~~~

 - When using Kubernetes, obtain a shell in a Sqedule pod:

   ~~~bash
   kubectl exec -ti deploy/sqedule -- sh
   ~~~

   The `sqedule-server` executable is in PATH so you can invoke from this shell:

   ~~~bash
   sqedule-server --help
   ~~~

   Note: the container environment is Alpine.

### Subcommands

To see what subcommands `sqedule-server` supports, run it with `--help`. You should see something like this:

~~~
$ sqedule-server --help
Sqedule server

Usage:
  sqedule-server [command]

Available Commands:
  db          Manage the database
  help        Help about any command
  run         Run the Sqedule API server

Flags:
      --config string      config file (default $HOME/.sqedule-server.yaml)
  -h, --help               help for sqedule-server
      --log-level string   log level, one of: error,warn,info,silent (default "info")

Use "sqedule-server [command] --help" for more information about a command.
~~~

## Configuration

### Ways to pass configuration

There are 3 ways to pass configuration to the Sqedule server:

 1. From the configuration file. Recommended when you've installed Sqedule via the binary (without containerization).

    The file is in YAML format. The Sqedule server looks for the configuration file in `~/.sqedule-server.yaml` by default. You can customize the location by running the Sqedule server with `--config /path-to-actual-file.yml`.

 2. From environment variables. Recommended when you're using containerization or Kubernetes.

 3. From CLI parameters.

    **Caveat**: see [Configuration naming format](#configuration-naming-format) to learn about the caveats of setting boolean options.

These methods are ordered from lowest to highest priority. CLI parameters override environment variables. Environment variables override the configuration file.

### Configuration naming format

Configuration options have a **canonical name**, for example `db-type`.

 - When using the configuration file, specify configuration options using their canonical names. Example: `db-type: something`
 - When using environment variables, specify configuration options with the `SQEDULE_` prefix, in uppercase, with dashes (`-`) replaced by underscores (`_`). Example: `SQEDULE_DB_TYPE=something`
 - When using CLI parameters, specify two dashes (`--`) followed by the canonical name. Example: `--db-type=something`

   **Caveat**: when setting a boolean option using a CLI parameter, be sure to specify the value after a `=`, like this: `--something=true_or_false`. Specifying the value after a space, like `--something true_or_false` doesn't work.

> **Note**: the rest of this document only specify configuration names in their canonical format.

### Configuration options

#### Global options

All `sqedule-server` subcommands accept these configuration options:

 * `log-level` (default: `info`) — Specifies the logging level for most messages. Must be one of: `error`, `warn`, `info` or `silent`.

#### `run` options

The `sqedule-server run` subcommand accepts these configuration options.

Database:

 * `db-type` (required) — The database type. Currently this can only be set to `postgresql`.
 * `db-connection` — Database connection string containing details such as location and credentials. The format is database-dependent. See [PostgreSQL connection string](#postgresql-connection-string).
 * `db-log-level` (default: `silent`) — Specifies the logging level for database activity messages. Must be one of: `error`, `warn`, `info` or `silent`.
 * `auto-db-migrate` (default: `true`) — Whether to automatically [migrate the database schema](#database-schema-migration) during startup.

HTTP server:

 * `bind`
 * `port`
 * `cors-origin`

### PostgreSQL connection string

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

## Security

The Sqedule HTTP server is **unprotected**. There is currently no built-in support for authentication. Therefore you should not expose it to the Internet directly. Instead, you should put it behind an HTTP reverse proxy that supports authentication, for example Nginx with HTTP basic authentication enabled.

Support for user accounts, and thus authentication, is planned for a future release.

## Database schema migration

During startup, the Sqedule server checks whether the database schema is up-to-date. If not then it will automatically migrate the database schema to the latest version.

Schema database migration is not concurrency-safe. This is the main reason why Sqedule currently [does not support running multiple concurrent instances](#multi-instance-safety) in its default configuration.

### Manual database schema migration

You can disable automatic database schema migration. If you do this, you are responsible for running database schema migrations manually every time you upgrade Sqedule.

To disable automatic migration, set `auto-db-migrate` to false.

To manually migrate the database schema, invoke the `sqedule-server db migrate` subcommand. This subcommand requires the `db-type` and `db-connection` configuration options.

## Multi-instance safety

In its default configuration, the Sqedule server does not support running multiple concurrent instances. The main reason for this is because it [automatically migrates the database schema during startup](#database-schema-migration), which is not concurrency-safe.

### Running multiple instances of Sqedule

To make it safe to run multiple concurrent instances of the Sqedule server, [disable automatic database schema migration](#manual-database-schema-migration).
