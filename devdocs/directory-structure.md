# Directory structure

## Server

Server entrypoint sources:

 * `cmd/sqedule-server/main.go` — Sqedule server's main function. It's deliberately minimal because it just kicks off subcommands with [Cobra](https://github.com/spf13/cobra) (similar to how the `git` command does nothing but kicking off subcommands).
 * `cmd/sqedule-server/*.go` — All server subcommands, backed by [Cobra](https://github.com/spf13/cobra).

The main server code is located in `server/`. Here are its subdirectories:

 * `approvalrulesengine/` — The engine for applying Approval Rules on Releases.
 * `authz` — Authorization. Determining which users are allowed to perform what actions on which domain resources.
 * `dbmigrations/` — Database migrations, backed by [Gormigrate](https://github.com/go-gormigrate/gormigrate).
 * `dbmodels/` — Database models, backed by [GORM](https://gorm.io/).
 * `dbutils/` — Generic database-related utility code that's not specific to Sqedule's business domain.
 * `httpapi/` — The backend JSON API server (backed by [Gin](https://github.com/gin-gonic/gin)) and everything closely related. Includes routing.
     - `auth/` — JSON API server authentication.
     - `controllers/` — Main code for handling API endpoints.
     - `json/` — Database-models-to-JSON serialization, and JSON-to-database-models conversion.

## Web UI

 * `webui/` — The web interface frontend. It's a SPA backed by [Next.js](https://nextjs.org/), [React](https://reactjs.org/) and [Material UI](https://material-ui.com/).

## Miscellaneous

 * `devdocs/` — Development documentation.
 * `devtools/` — Tools and scripts used during development.
 * `bin/` - Place to build binaries to in automation
