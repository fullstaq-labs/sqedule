package main

import (
	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json/proposalstateinput"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// applicationApprovalRulesetBindingCmd represents the 'application-approval-ruleset-binding' command
var applicationApprovalRulesetBindingCmd = &cobra.Command{
	Use:   "application-approval-ruleset-binding",
	Short: "Manage application approval ruleset bindings",
}

func defineApplicationApprovalRulesetBindingCreateOrUpdateFlags(flags *pflag.FlagSet) {
	flags.String("mode", "enforcing", "binding mode; 'enforcing' or 'permissive'")
	flags.String("proposal-state", "draft", "'draft', 'final' or 'abandon'")
	flags.Bool("enabled", true, "whether to enable this application ruleset binding")
}

func applicationApprovalRulesetBindingCreateOrUpdateCmd_createVersionInput(viper *viper.Viper, creating bool) json.ApplicationApprovalRulesetBindingVersionInput {
	var mode *approvalrulesetbindingmode.Mode

	if creating {
		modeVal := approvalrulesetbindingmode.Mode(viper.GetString("mode"))
		mode = &modeVal
	} else if modeStr := cli.GetViperStringIfSet(viper, "mode"); modeStr != nil {
		modeVal := approvalrulesetbindingmode.Mode(*modeStr)
		mode = &modeVal
	}

	return json.ApplicationApprovalRulesetBindingVersionInput{
		ReviewableVersionInputBase: json.ReviewableVersionInputBase{
			ProposalState: proposalstateinput.Input(viper.GetString("proposal-state")),
		},
		Mode:    mode,
		Enabled: cli.GetViperBoolIfSet(viper, "enabled"),
	}
}

func init() {
	rootCmd.AddCommand(applicationApprovalRulesetBindingCmd)
}
