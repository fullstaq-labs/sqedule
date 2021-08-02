package main

import (
	encjson "encoding/json"
	"fmt"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applicationListCmd represents the 'application list' command
var applicationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List applications",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return applicationListCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationListCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	config := cli.LoadConfigFromViper(viper)
	state, err := cli.LoadStateFromFilesystem()
	if err != nil {
		return fmt.Errorf("Error loading state: %w", err)
	}

	req, err := cli.NewApiRequest(config, state)
	if err != nil {
		return err
	}

	var result interface{}
	resp, err := req.
		SetResult(&result).
		Get("/applications")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error listing applications: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))

	return nil
}

func init() {
	cmd := applicationListCmd
	flags := cmd.Flags()
	applicationCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)
}
