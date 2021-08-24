package main

import (
	"fmt"

	"github.com/fullstaq-labs/sqedule/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// versionCmd represents the 'version' command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show server version",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		fmt.Println(server.VersionString)
		return nil
	},
}

func init() {
	cmd := versionCmd
	rootCmd.AddCommand(cmd)
}
