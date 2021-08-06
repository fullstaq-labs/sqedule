package main

import (
	encjson "encoding/json"
	"fmt"
	"net/url"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applicationProposalCreateCmd represents the 'application proposal create' command
var applicationProposalCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an application proposal",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return applicationProposalCreateCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationProposalCreateCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := applicationProposalCreateCmd_checkConfig(viper)
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
		SetBody(applicationProposalCreateCmd_createBody(viper)).
		SetResult(&result).
		Patch(fmt.Sprintf("/applications/%s",
			url.PathEscape(viper.GetString("application-id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error creating application proposal: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))
	cli.PrintSeparatorln(printer)
	cli.PrintCelebrationlnf(printer, "Approval ruleset proposal (ID=%v) created!",
		applicationProposalCreateCmd_getProposalID(result))

	return nil
}

func applicationProposalCreateCmd_createBody(viper *viper.Viper) map[string]interface{} {
	return map[string]interface{}{
		"version": applicationCreateOrUpdateCmd_createVersionInput(viper),
	}
}

func applicationProposalCreateCmd_getProposalID(resource map[string]interface{}) interface{} {
	version := resource["version"].(map[string]interface{})
	return version["id"]
}

func applicationProposalCreateCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"application-id"},
	})
}

func init() {
	cmd := applicationProposalCreateCmd
	flags := cmd.Flags()
	applicationProposalCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("application-id", "", "Application ID (required)")
	defineApplicationCreateOrUpdateFlags(flags, false)
}
