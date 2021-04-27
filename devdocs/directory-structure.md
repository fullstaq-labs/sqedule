# Directory structure

Basic structure:

 * `main.go` — Sqedule's main program entrypoint. It's deliberately minimal because it just kicks off subcommands with [Cobra](https://github.com/spf13/cobra) (similar to how the `git` command does nothing but kicking off subcommands).
 * `cmd/` — All subcommands, backed by [Cobra](https://github.com/spf13/cobra).

Core code:

 * `approvalrulesengine/` — The engine for applying Approval Rules on Releases.
 * `dbmigrations/` — Database migrations, backed by [Gormigrate](https://github.com/go-gormigrate/gormigrate).
 * `dbmodels/` — Database models, backed by [GORM](https://gorm.io/).
 * `dbutils/` — Generic database-related utility code that's not specific to Sqedule's business domain.
 * `httpapi/` — The backend JSON API server (backed by [Gin](https://github.com/gin-gonic/gin)) and everything closely related. Routing, authentication, authorization, database-to-JSON serialization, and JSON-to-database conversion.
 * `webui/` — The web interface frontend. It's a SPA backed by [Next.js](https://nextjs.org/), [React](https://reactjs.org/) and [Material UI](https://material-ui.com/).

Miscellaneous:

 * `devdocs/` — Development documentation.
 * `devtools/` — Tools and scripts used during development.
