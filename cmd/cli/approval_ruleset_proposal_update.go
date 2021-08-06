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

// approvalRulesetProposalUpdateCmd represents the 'approval-ruleset proposal update' command
var approvalRulesetProposalUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an approval ruleset proposal",
	Long:  "Update an approval ruleset proposal's properties, but not its rules. To manage rules, use `sqedule approval-ruleset proposal rule`",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return approvalRulesetProposalUpdateCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func approvalRulesetProposalUpdateCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := approvalRulesetProposalUpdateCmd_checkConfig(viper)
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
		SetBody(approvalRulesetCreateOrUpdateCmd_createVersionInput(viper)).
		SetResult(&result).
		Patch(fmt.Sprintf("/approval-rulesets/%s/proposals/%s",
			url.PathEscape(viper.GetString("approval-ruleset-id")),
			url.PathEscape(viper.GetString("id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error updating approval ruleset proposal: %s", cli.GetApiErrorMessage(resp))
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

func approvalRulesetProposalUpdateCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"approval-ruleset-id", "id"},
	})
}

func init() {
	cmd := approvalRulesetProposalUpdateCmd
	flags := cmd.Flags()
	approvalRulesetProposalCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("approval-ruleset-id", "", "Approval ruleset ID (required)")
	flags.String("id", "", "Proposal ID (required)")
	defineApprovalRulesetCreateOrUpdateFlags(flags, false)
}
