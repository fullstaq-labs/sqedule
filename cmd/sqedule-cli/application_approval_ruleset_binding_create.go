package main

import (
	encjson "encoding/json"
	"fmt"
	"net/url"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applicationApprovalRulesetBindingCreateCmd represents the 'approval-ruleset create' command
var applicationApprovalRulesetBindingCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an application application approval ruleset binding binding",
	Long:  "Creates an application application approval ruleset binding binding",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return applicationApprovalRulesetBindingCreateCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationApprovalRulesetBindingCreateCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := applicationApprovalRulesetBindingCreateCmd_checkConfig(viper)
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

	var ruleset map[string]interface{}
	resp, err := req.
		SetBody(applicationApprovalRulesetBindingCreateCmd_createBody(viper)).
		SetResult(&ruleset).
		Post(fmt.Sprintf("/applications/%s/approval-ruleset-bindings/%s",
			url.PathEscape(viper.GetString("application-id")),
			url.PathEscape(viper.GetString("approval-ruleset-id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error creating application application approval ruleset binding binding: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(ruleset, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))
	cli.PrintSeparatorln(printer)
	cli.PrintCelebrationlnf(printer, "Application application approval ruleset binding binding created!")

	version := ruleset["version"].(map[string]interface{})
	if version["version_state"] == "approved" {
		cli.PrintTiplnf(printer, "It has been auto-approved by the system. To view it, use `sqedule application-approval-ruleset-binding describe`")
	} else {
		cli.PrintTiplnf(printer, "It is still a proposal. To view it, use `sqedule application-approval-ruleset-binding proposal list`")
	}

	return nil
}

func applicationApprovalRulesetBindingCreateCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"application-id", "approval-ruleset-id"},
	})
}

func applicationApprovalRulesetBindingCreateCmd_createBody(viper *viper.Viper) json.ApplicationApprovalRulesetBindingInput {
	version := applicationApprovalRulesetBindingCreateOrUpdateCmd_createVersionInput(viper, true)
	return json.ApplicationApprovalRulesetBindingInput{
		Version: &version,
	}
}

func init() {
	cmd := applicationApprovalRulesetBindingCreateCmd
	flags := cmd.Flags()
	applicationApprovalRulesetBindingCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("application-id", "", "ID of the application to bind to (required)")
	flags.String("approval-ruleset-id", "", "ID of the approval ruleset to bind to (required)")
	defineApplicationApprovalRulesetBindingCreateOrUpdateFlags(flags, true)
}
