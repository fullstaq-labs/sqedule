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

// applicationApprovalRulesetBindingProposalCreateCmd represents the 'application-approval-ruleset-binding proposal create' command
var applicationApprovalRulesetBindingProposalCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an application approval ruleset binding proposal",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			return err
		}

		return applicationApprovalRulesetBindingProposalCreateCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationApprovalRulesetBindingProposalCreateCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := applicationApprovalRulesetBindingProposalCreateCmd_checkConfig(viper)
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

	var result map[string]interface{}
	resp, err := req.
		SetBody(applicationApprovalRulesetBindingProposalCreateCmd_createBody(viper)).
		SetResult(&result).
		Patch(fmt.Sprintf("/application-approval-ruleset-bindings/%s/%s",
			url.PathEscape(viper.GetString("application-id")),
			url.PathEscape(viper.GetString("approval-ruleset-id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error creating application approval ruleset binding proposal: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))
	cli.PrintSeparatorln(printer)
	cli.PrintCelebrationlnf(printer, "Approval ruleset proposal (ID=%v) created!",
		applicationApprovalRulesetBindingProposalCreateCmd_getProposalID(result))

	return nil
}

func applicationApprovalRulesetBindingProposalCreateCmd_createBody(viper *viper.Viper) map[string]interface{} {
	return map[string]interface{}{
		"version": applicationApprovalRulesetBindingCreateOrUpdateCmd_createVersionInput(viper, true),
	}
}

func applicationApprovalRulesetBindingProposalCreateCmd_getProposalID(ruleset map[string]interface{}) interface{} {
	version := ruleset["version"].(map[string]interface{})
	return version["id"]
}

func applicationApprovalRulesetBindingProposalCreateCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"application-id", "approval-ruleset-id"},
	})
}

func init() {
	cmd := applicationApprovalRulesetBindingProposalCreateCmd
	flags := cmd.Flags()
	applicationApprovalRulesetBindingProposalCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("application-id", "", "ID of the bound application (required)")
	flags.String("approval-ruleset-id", "", "ID of the bound application approval ruleset (required)")
	defineApplicationApprovalRulesetBindingCreateOrUpdateFlags(flags)
}
