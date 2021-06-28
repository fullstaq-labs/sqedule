package main

import (
	"github.com/spf13/cobra"
)

// approvalRulesetVersionCmd represents the 'approval-ruleset version' command
var approvalRulesetVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage approval rulesets versions",
}

func init() {
	approvalRulesetCmd.AddCommand(approvalRulesetVersionCmd)
}
