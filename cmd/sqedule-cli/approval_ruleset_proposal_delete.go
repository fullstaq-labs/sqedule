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

// approvalRulesetProposalDeleteCmd represents the 'approval-ruleset proposal delete' command
var approvalRulesetProposalDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an approval ruleset proposal",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return approvalRulesetProposalDeleteCmd_run(viper.GetViper(), mocking.RealPrinter{}, false)
	},
}

func approvalRulesetProposalDeleteCmd_run(viper *viper.Viper, printer mocking.IPrinter, testing bool) error {
	err := approvalRulesetProposalDeleteCmd_checkConfig(viper)
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
		SetResult(&result).
		Delete(fmt.Sprintf("/approval-rulesets/%s/proposals/%s",
			url.PathEscape(viper.GetString("approval-ruleset-id")),
			url.PathEscape(viper.GetString("id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error deleting approval ruleset proposal: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))

	return nil
}

func approvalRulesetProposalDeleteCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"approval-ruleset-id", "id"},
	})
}

func init() {
	cmd := approvalRulesetProposalDeleteCmd
	flags := cmd.Flags()
	approvalRulesetProposalCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("approval-ruleset-id", "", "Approval ruleset ID (required)")
	flags.String("id", "", "Proposal ID (required)")
}
