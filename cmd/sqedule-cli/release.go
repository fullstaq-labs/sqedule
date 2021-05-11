package main

import (
	"github.com/spf13/cobra"
)

// releaseCmd represents the 'release' command
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Manage releases",
}

func init() {
	rootCmd.AddCommand(releaseCmd)
}
