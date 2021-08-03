package main

import (
	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json/proposalstate"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// applicationCmd represents the 'application' command
var applicationCmd = &cobra.Command{
	Use:   "application",
	Short: "Manage applications",
}

func defineApplicationCreateOrUpdateFlags(flags *pflag.FlagSet, creating bool) {
	var requiredAtCreationIndicator string
	if creating {
		requiredAtCreationIndicator = " (required)"
	}

	flags.String("display-name", "", "Human-friendly display name"+requiredAtCreationIndicator)
	flags.String("proposal-state", "draft", "'draft', 'final' or 'abandon'")
	flags.Bool("enabled", true, "Whether to enable this ruleset")
}

func applicationCreateOrUpdateCmd_createVersionInput(viper *viper.Viper) json.ApplicationVersionInput {
	return json.ApplicationVersionInput{
		ReviewableVersionInputBase: json.ReviewableVersionInputBase{
			ProposalState: proposalstate.State(viper.GetString("proposal-state")),
		},
		DisplayName: cli.GetViperStringIfSet(viper, "display-name"),
		Enabled:     cli.GetViperBoolIfSet(viper, "enabled"),
	}
}

func init() {
	rootCmd.AddCommand(applicationCmd)
}
