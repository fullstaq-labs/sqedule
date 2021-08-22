# Configuration reference

## Global options

All [`sqedule-server` subcommands](../concepts/server-exe.md) accept these configuration options:

 * `log-level` (default: `info`) — Specifies the logging level for most messages. Must be one of: `error`, `warn`, `info` or `silent`.

## `run` options

The `sqedule-server run` [subcommand](../concepts/server-exe.md) accepts these configuration options.

### Database

 * `db-type` (required) — The database type. Currently this can only be set to `postgresql`.
 * `db-connection` (string) — Database connection string containing details such as location and credentials. The format is database-dependent. See [PostgreSQL connection string](#postgresql-connection-string).
 * `db-log-level` (default: `silent`) — Specifies the logging level for database activity messages. Must be one of: `error`, `warn`, `info` or `silent`.
 * `auto-db-migrate` (boolean, default: `true`) — Whether to automatically [migrate the database schema](#database-schema-migration) during startup.

### HTTP server

 * `bind` (string, default: `localhost`) — The IP/hostname to bind on.
 * `port` (integer, default: `3001`) — The port to bind on.
 * `cors-origin` (string) — Allow requests from the given CORS origin (e.g. `https://yourhost.com`). Commands Sqedule to output CORS preflight responses that allow this origin.
