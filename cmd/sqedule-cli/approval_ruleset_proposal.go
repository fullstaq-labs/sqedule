package main

import (
	"github.com/spf13/cobra"
)

// approvalRulesetProposalCmd represents the 'approval-ruleset proposal' command
var approvalRulesetProposalCmd = &cobra.Command{
	Use:   "proposal",
	Short: "Manage proposals",
}

func init() {
	approvalRulesetCmd.AddCommand(approvalRulesetProposalCmd)
}
