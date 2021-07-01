package main

import (
	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json/proposalstate"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// approvalRulesetCmd represents the 'approval-ruleset' command
var approvalRulesetCmd = &cobra.Command{
	Use:   "approval-ruleset",
	Short: "Manage approval rulesets",
}

func defineApprovalRulesetCreateOrUpdateFlags(flags *pflag.FlagSet) {
	flags.String("display-name", "", "Human-friendly display name")
	flags.String("description", "", "")
	flags.String("proposal-state", "draft", "'draft', 'final' or 'abandon'")
	flags.Bool("enabled", true, "Whether to enable this ruleset")
}

func approvalRulesetCreateOrUpdateCmd_createVersionInput(viper *viper.Viper) json.ApprovalRulesetVersionInput {
	return json.ApprovalRulesetVersionInput{
		ReviewableVersionInputBase: json.ReviewableVersionInputBase{
			ProposalState: proposalstate.State(viper.GetString("proposal-state")),
		},
		DisplayName: cli.GetViperStringIfSet(viper, "display-name"),
		Description: cli.GetViperStringIfSet(viper, "description"),
		Enabled:     cli.GetViperBoolIfSet(viper, "enabled"),
	}
}

func init() {
	rootCmd.AddCommand(approvalRulesetCmd)
}
