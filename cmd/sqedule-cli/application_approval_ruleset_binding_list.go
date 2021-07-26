package main

import (
	encjson "encoding/json"
	"fmt"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applicationApprovalRulesetBindingListCmd represents the 'application-approval-ruleset-binding list' command
var applicationApprovalRulesetBindingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List application approval ruleset bindings",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return applicationApprovalRulesetBindingListCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationApprovalRulesetBindingListCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	config := cli.LoadConfigFromViper(viper)
	state, err := cli.LoadStateFromFilesystem()
	if err != nil {
		return fmt.Errorf("Error loading state: %w", err)
	}

	req, err := cli.NewApiRequest(config, state)
	if err != nil {
		return err
	}

	var result interface{}
	resp, err := req.
		SetResult(&result).
		Get("/application-approval-ruleset-bindings")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error listing application approval ruleset bindings: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))

	return nil
}

func init() {
	cmd := applicationApprovalRulesetBindingListCmd
	flags := cmd.Flags()
	applicationApprovalRulesetBindingCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)
}
