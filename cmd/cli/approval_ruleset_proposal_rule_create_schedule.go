package main

import (
	encjson "encoding/json"
	"fmt"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// approvalRulesetProposalRuleCreateScheduleCmd represents the 'approval-ruleset proposal rule create-schedule' command
var approvalRulesetProposalRuleCreateScheduleCmd = &cobra.Command{
	Use:   "create-schedule",
	Short: "Create a schedule rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return approvalRulesetProposalRuleCreateScheduleCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func approvalRulesetProposalRuleCreateScheduleCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := approvalRulesetProposalRuleCreateScheduleCmd_checkConfig(viper)
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
	rules = append(rules, approvalRulesetProposalRuleCreateScheduleCmd_buildRuleDefinition(viper))
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
	cli.PrintCelebrationlnf(printer, "Rule created!")

	return nil
}

func approvalRulesetProposalRuleCreateScheduleCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"approval-ruleset-id", "proposal-id"},
	})
}

func approvalRulesetProposalRuleCreateScheduleCmd_buildRuleDefinition(viper *viper.Viper) map[string]interface{} {
	result := map[string]interface{}{
		"type":    "schedule",
		"enabled": viper.GetBool("enabled"),
	}

	maybeAddString := func(resultKey string, viperKey string) {
		value := viper.GetString(viperKey)
		if len(value) > 0 {
			result[resultKey] = value
		}
	}

	maybeAddString("begin_time", "begin-time")
	maybeAddString("end_time", "end-time")
	maybeAddString("days_of_week", "days-of-week")
	maybeAddString("days_of_month", "days-of-month")
	maybeAddString("months-of-year", "months-of-year")

	return result
}

func init() {
	cmd := approvalRulesetProposalRuleCreateScheduleCmd
	flags := cmd.Flags()
	approvalRulesetProposalRuleCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("approval-ruleset-id", "", "Approval ruleset ID (required)")
	flags.String("proposal-id", "", "Proposal ID (required)")
	flags.Bool("enabled", true, "Whether to enable this rule")
	flags.String("begin-time", "", "Schedule begin time")
	flags.String("end-time", "", "Schedule end time")
	flags.String("days-of-week", "", "Schedule days of week")
	flags.String("days-of-month", "", "Schedule days of month")
	flags.String("months-of-year", "", "Schedule months of year")
}
