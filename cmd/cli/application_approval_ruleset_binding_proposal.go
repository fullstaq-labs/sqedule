package main

import (
	"github.com/spf13/cobra"
)

// applicationApprovalRulesetBindingProposalCmd represents the 'application-approval-ruleset-binding proposal' command
var applicationApprovalRulesetBindingProposalCmd = &cobra.Command{
	Use:   "proposal",
	Short: "Manage proposals",
}

func init() {
	applicationApprovalRulesetBindingCmd.AddCommand(applicationApprovalRulesetBindingProposalCmd)
}
