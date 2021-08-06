package main

import (
	"github.com/spf13/cobra"
)

// applicationProposalCmd represents the 'application proposal' command
var applicationProposalCmd = &cobra.Command{
	Use:   "proposal",
	Short: "Manage proposals",
}

func init() {
	applicationCmd.AddCommand(applicationProposalCmd)
}
