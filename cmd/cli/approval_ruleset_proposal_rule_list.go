package main

import (
	encjson "encoding/json"
	"fmt"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// approvalRulesetProposalRuleListCmd represents the 'approval-ruleset proposal rule list' command
var approvalRulesetProposalRuleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			return err
		}

		return approvalRulesetProposalRuleListCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func approvalRulesetProposalRuleListCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := approvalRulesetProposalRuleListCmd_checkConfig(viper)
	if err != nil {
		return err
	}

	config := cli.LoadConfigFromViper(viper)
	state, err := cli.LoadStateFromFilesystem()
	if err != nil {
		return fmt.Errorf("Error loading state: %w", err)
	}

	ruleset, err := approvalRulesetProposalCmd_getRuleset(viper, config, state)
	if err != nil {
		return err
	}
	rules, err := approvalRulesetProposalRuleCmd_getRules(ruleset)
	if err != nil {
		return err
	}

	output, err := encjson.MarshalIndent(rules, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))

	return nil
}

func approvalRulesetProposalRuleListCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"approval-ruleset-id", "proposal-id"},
	})
}

func init() {
	cmd := approvalRulesetProposalRuleListCmd
	flags := cmd.Flags()
	approvalRulesetProposalRuleCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("approval-ruleset-id", "", "approval ruleset ID (required)")
	flags.String("proposal-id", "", "proposal ID (required)")
}
