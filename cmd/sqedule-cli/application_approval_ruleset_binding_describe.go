package main

import (
	encjson "encoding/json"
	"fmt"
	"net/url"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applicationApprovalRulesetBindingDescribeCmd represents the 'approval-ruleset describe' command
var applicationApprovalRulesetBindingDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe an approval ruleset",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return applicationApprovalRulesetBindingDescribeCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationApprovalRulesetBindingDescribeCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := applicationApprovalRulesetBindingDescribeCmd_checkConfig(viper)
	if err != nil {
		return err
	}

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
		Get(fmt.Sprintf("/application-approval-ruleset-bindings/%s/%s",
			url.PathEscape(viper.GetString("application-id")),
			url.PathEscape(viper.GetString("approval-ruleset-id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error describing application approval ruleset binding: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))

	return nil
}

func applicationApprovalRulesetBindingDescribeCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"application-id", "approval-ruleset-id"},
	})
}

func init() {
	cmd := applicationApprovalRulesetBindingDescribeCmd
	flags := cmd.Flags()
	applicationApprovalRulesetBindingCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("application-id", "", "ID of the bound application (required)")
	flags.String("approval-ruleset-id", "", "ID of the bound application approval ruleset (required)")
}
