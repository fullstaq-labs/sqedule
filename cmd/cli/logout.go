package main

import (
	"fmt"
	"time"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/spf13/cobra"
)

// logoutCmd represents the 'login' command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Discards the Sqedule authorization token",
	RunE:  LogoutCmd_Run,
}

func LogoutCmd_Run(cmd *cobra.Command, args []string) error {
	state, err := cli.LoadStateFromFilesystem()
	if err != nil {
		return fmt.Errorf("Error loading state: %w", err)
	}

	state.AuthToken = ""
	state.AuthTokenExpirationTime = time.Time{}
	err = state.SaveToFilesystem()
	if err != nil {
		return fmt.Errorf("Error saving state: %w", err)
	}

	return nil
}

func init() {
	cmd := logoutCmd
	rootCmd.AddCommand(cmd)
}
