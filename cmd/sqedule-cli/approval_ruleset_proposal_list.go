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

// approvalRulesetProposalListCmd represents the 'approval-ruleset proposal list' command
var approvalRulesetProposalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List approval ruleset versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return approvalRulesetProposalListCmd_run(viper.GetViper(), mocking.RealPrinter{}, false)
	},
}

func approvalRulesetProposalListCmd_run(viper *viper.Viper, printer mocking.IPrinter, testing bool) error {
	err := approvalRulesetProposalListCmd_checkConfig(viper)
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
		Get(fmt.Sprintf("/approval-rulesets/%s/proposals",
			url.PathEscape(viper.GetString("approval-ruleset-id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error listing approval ruleset proposals: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.Println(string(output))

	return nil
}

func approvalRulesetProposalListCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"approval-ruleset-id"},
	})
}

func init() {
	cmd := approvalRulesetProposalListCmd
	flags := cmd.Flags()
	approvalRulesetProposalCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("approval-ruleset-id", "", "")
}
