package main

import (
	"github.com/spf13/cobra"
)

// approvalRulesetProposalRuleCmd represents the 'approval-ruleset proposal rule' command
var approvalRulesetProposalRuleCmd = &cobra.Command{
	Use:   "rule",
	Short: "Manage rules",
}

func init() {
	approvalRulesetProposalCmd.AddCommand(approvalRulesetProposalRuleCmd)
}
