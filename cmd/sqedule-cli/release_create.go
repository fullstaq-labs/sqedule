package main

import (
	"github.com/spf13/cobra"
)

// releaseCreateCmd represents the 'release create' command
var releaseCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a release",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	cmd := releaseCreateCmd
	//flags := cmd.Flags()
	releaseCmd.AddCommand(cmd)
}
