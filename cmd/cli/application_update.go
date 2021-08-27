package main

import (
	encjson "encoding/json"
	"fmt"
	"net/url"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applicationUpdateCmd represents the 'application update' command
var applicationUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an application's non-versioned properties",
	// editorconfig-checker-disable
	Long: `Update an application's non-versioned properties.

To update its versioned properties (e.g. display name or rules):

 1. create a proposal first: ` + "`sqedule application proposal create`" + `
 2. then use ` + "`sqedule application proposal update`" + ` and ` + "`sqedule application proposal rule`",
	// editorconfig-checker-enable
	RunE: func(cmd *cobra.Command, args []string) error {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			return err
		}

		return applicationUpdateCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationUpdateCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := applicationUpdateCmd_checkConfig(viper)
	if err != nil {
		return err
	}

	config := cli.LoadConfigFromViper(viper)
	state, err := cli.LoadStateFromFilesystem()
	if err != nil {
		return fmt.Errorf("Error loading state: %w", err)
	}

	req, err := cli.NewApiRequest(config, state)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	resp, err := req.
		SetBody(applicationUpdateCmd_createBody(viper)).
		SetResult(&result).
		Patch(fmt.Sprintf("/applications/%s",
			url.PathEscape(viper.GetString("id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error updating application: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))
	cli.PrintSeparatorln(printer)
	cli.PrintCelebrationlnf(printer, "Application '%s' updated!", viper.GetString("id"))

	return nil
}

func applicationUpdateCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"id"},
	})
}

func applicationUpdateCmd_createBody(viper *viper.Viper) json.ApplicationInput {
	return json.ApplicationInput{
		ID: cli.GetViperStringIfSet(viper, "set-id"),
	}
}

func init() {
	cmd := applicationUpdateCmd
	flags := cmd.Flags()
	applicationCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("id", "", "ID of application to update (required)")
	flags.String("set-id", "", "change application ID")
}
