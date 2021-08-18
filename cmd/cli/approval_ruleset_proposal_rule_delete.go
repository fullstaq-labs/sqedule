package main

import (
	encjson "encoding/json"
	"fmt"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// approvalRulesetProposalRuleDeleteCmd represents the 'approval-ruleset proposal rule delete' command
var approvalRulesetProposalRuleDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return approvalRulesetProposalRuleDeleteCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func approvalRulesetProposalRuleDeleteCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := approvalRulesetProposalRuleDeleteCmd_checkConfig(viper)
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
	rules = approvalRulesetProposalRuleDeleteCmd_removeRule(viper, rules)
	ruleset, err = approvalRulesetProposalRuleCmd_patchRules(viper, config, state, rules)
	if err != nil {
		return err
	}
	rules, err = approvalRulesetProposalRuleCmd_getRules(ruleset)
	if err != nil {
		return err
	}

	output, err := encjson.MarshalIndent(rules, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))
	cli.PrintSeparatorln(printer)
	cli.PrintCelebrationlnf(printer, "Rule deleted!")

	return nil
}

func approvalRulesetProposalRuleDeleteCmd_removeRule(viper *viper.Viper, rules []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rules))
	targetRuleID := viper.GetString("id")

	for _, rule := range rules {
		var retain = true
		if ruleID, ok := rule["id"]; ok {
			ruleIDString := fmt.Sprintf("%v", ruleID)
			if ruleIDString == targetRuleID {
				retain = false
			}
		}

		if retain {
			result = append(result, rule)
		}
	}

	return result
}

func approvalRulesetProposalRuleDeleteCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"approval-ruleset-id", "proposal-id", "id"},
	})
}

func init() {
	cmd := approvalRulesetProposalRuleDeleteCmd
	flags := cmd.Flags()
	approvalRulesetProposalRuleCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("approval-ruleset-id", "", "approval ruleset ID (required)")
	flags.String("proposal-id", "", "proposal ID (required)")
	flags.String("id", "", "rule ID (required)")
}
