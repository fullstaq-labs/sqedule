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

// applicationProposalUpdateCmd represents the 'application proposal update' command
var applicationProposalUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an application proposal",
	Long:  "Update an application proposal's properties, but not its rules. To manage rules, use `sqedule application proposal rule`",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return applicationProposalUpdateCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationProposalUpdateCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := applicationProposalUpdateCmd_checkConfig(viper)
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

	var result interface{}
	resp, err := req.
		SetBody(applicationCreateOrUpdateCmd_createVersionInput(viper)).
		SetResult(&result).
		Patch(fmt.Sprintf("/applications/%s/proposals/%s",
			url.PathEscape(viper.GetString("application-id")),
			url.PathEscape(viper.GetString("id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error updating application proposal: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))
	cli.PrintSeparatorln(printer)
	cli.PrintCelebrationlnf(printer, "Proposal updated!")

	return nil
}

func applicationProposalUpdateCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"application-id", "id"},
	})
}

func init() {
	cmd := applicationProposalUpdateCmd
	flags := cmd.Flags()
	applicationProposalCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("application-id", "", "Application ID (required)")
	flags.String("id", "", "Proposal ID (required)")
	defineApplicationCreateOrUpdateFlags(flags, false)
}
