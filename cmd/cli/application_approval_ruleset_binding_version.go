package main

import (
	"github.com/spf13/cobra"
)

// applicationApprovalRulesetBindingVersionCmd represents the 'application-approval-ruleset-binding version' command
var applicationApprovalRulesetBindingVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage approved versions",
}

func init() {
	applicationApprovalRulesetBindingCmd.AddCommand(applicationApprovalRulesetBindingVersionCmd)
}
