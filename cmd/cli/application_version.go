package main

import (
	"github.com/spf13/cobra"
)

// applicationVersionCmd represents the 'application version' command
var applicationVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage approved versions",
}

func init() {
	applicationCmd.AddCommand(applicationVersionCmd)
}
