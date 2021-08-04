package main

import (
	encjson "encoding/json"
	"fmt"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applicationCreateCmd represents the 'application create' command
var applicationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an application",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return applicationCreateCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationCreateCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := applicationCreateCmd_checkConfig(viper)
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
		SetBody(applicationCreateCmd_createBody(viper)).
		SetResult(&result).
		Post("/applications")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error creating application: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))
	cli.PrintSeparatorln(printer)
	cli.PrintCelebrationlnf(printer, "Application created!")

	version := result["version"].(map[string]interface{})
	if version["version_state"] == "approved" {
		cli.PrintTiplnf(printer, "It has been auto-approved by the system. To view it, use `sqedule application describe`")
	} else {
		cli.PrintTiplnf(printer, "It is still a proposal. To view it, use `sqedule application proposal list`")
	}

	return nil
}

func applicationCreateCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"id", "display-name"},
	})
}

func applicationCreateCmd_createBody(viper *viper.Viper) json.ApplicationInput {
	version := applicationCreateOrUpdateCmd_createVersionInput(viper)
	return json.ApplicationInput{
		ID:      lib.NewStringPtr(viper.GetString("id")),
		Version: &version,
	}
}

func init() {
	cmd := applicationCreateCmd
	flags := cmd.Flags()
	applicationCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("id", "", "A machine-friendly identifier (required)")
	defineApplicationCreateOrUpdateFlags(flags, true)
}
