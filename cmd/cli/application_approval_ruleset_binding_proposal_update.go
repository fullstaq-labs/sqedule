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

// applicationApprovalRulesetBindingProposalUpdateCmd represents the 'application-approval-ruleset-binding proposal update' command
var applicationApprovalRulesetBindingProposalUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an application approval ruleset binding proposal",
	Long:  "Update an application approval ruleset binding proposal's properties, but not its rules. To manage rules, use `sqedule application-approval-ruleset-binding proposal rule`",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			return err
		}

		return applicationApprovalRulesetBindingProposalUpdateCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationApprovalRulesetBindingProposalUpdateCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := applicationApprovalRulesetBindingProposalUpdateCmd_checkConfig(viper)
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
		SetBody(applicationApprovalRulesetBindingCreateOrUpdateCmd_createVersionInput(viper, false)).
		SetResult(&result).
		Patch(fmt.Sprintf("/application-approval-ruleset-bindings/%s/%s/proposals/%s",
			url.PathEscape(viper.GetString("application-id")),
			url.PathEscape(viper.GetString("approval-ruleset-id")),
			url.PathEscape(viper.GetString("id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error updating application approval ruleset binding proposal: %s", cli.GetApiErrorMessage(resp))
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

func applicationApprovalRulesetBindingProposalUpdateCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"application-id", "approval-ruleset-id", "id"},
	})
}

func init() {
	cmd := applicationApprovalRulesetBindingProposalUpdateCmd
	flags := cmd.Flags()
	applicationApprovalRulesetBindingProposalCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("application-id", "", "ID of the bound application (required)")
	flags.String("approval-ruleset-id", "", "ID of the bound application approval ruleset (required)")
	flags.String("id", "", "proposal ID (required)")
	defineApplicationApprovalRulesetBindingCreateOrUpdateFlags(flags)
}
