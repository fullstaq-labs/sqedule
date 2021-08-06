package main

import (
	"os"

	"github.com/gookit/color"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// dbHelpCmd represents the 'db help' command
var dbHelpCmd = &cobra.Command{
	Use:   "help",
	Short: "Help with supported database types and connection strings",
	Run: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())

		if !isatty.IsTerminal(os.Stdout.Fd()) {
			color.Disable()
		}

		// Source of this information: https://godoc.org/github.com/jackc/pgconn#ParseConfig
		// The GORM PostgreSQL driver uses pgx under the hood, in which turn
		// uses pgconn under the hood.
		color.Println(`<fg=white;bg=blue;op=bold>### Supported database types ###</>

The only supported database type right now is 'postgresql'.

  --db-type=postgresql

<fg=white;bg=blue;op=bold>### PostgreSQL connections ###</>

Connection strings can be in DSN format or in URL format. Examples:

  --db-connection='user=jack password=secret host=pg.example.com port=5432 dbname=mydb sslmode=verify-ca'
  --db-connection='postgres://jack:secret@pg.example.com:5432/mydb?sslmode=verify-ca'

Sqedule also reads from PG* environment variables, and (if a password is not provided) from the .pgpass file. Therefore, providing a connection string is optional.

Please refer the PostgreSQL documentation on:

 - Supported options in the connection string:
   https://www.postgresql.org/docs/11/libpq-connect.html#LIBPQ-PARAMKEYWORDS
 - Meaning of PG* environment variables:
   http://www.postgresql.org/docs/11/static/libpq-envars.html

Note that Sqedule only supports these environment variables:

  PGHOST
  PGPORT
  PGDATABASE
  PGUSER
  PGPASSWORD
  PGPASSFILE
  PGSERVICE
  PGSERVICEFILE
  PGSSLMODE
  PGSSLCERT
  PGSSLKEY
  PGSSLROOTCERT
  PGAPPNAME
  PGCONNECT_TIMEOUT
  PGTARGETSESSIONATTRS`)
	},
}

func init() {
	cmd := dbHelpCmd
	dbCmd.AddCommand(cmd)
}
