package main

import (
	"github.com/spf13/cobra"
)

// approvalRulesetCmd represents the 'approval-ruleset' command
var approvalRulesetCmd = &cobra.Command{
	Use:   "approval-ruleset",
	Short: "Manage approval rulesets",
}

func init() {
	rootCmd.AddCommand(approvalRulesetCmd)
}
