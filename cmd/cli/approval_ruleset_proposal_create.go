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

// approvalRulesetProposalCreateCmd represents the 'approval-ruleset proposal create' command
var approvalRulesetProposalCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an approval ruleset proposal",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			return err
		}

		return approvalRulesetProposalCreateCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func approvalRulesetProposalCreateCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := approvalRulesetProposalCreateCmd_checkConfig(viper)
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
		SetBody(approvalRulesetProposalCreateCmd_createBody()).
		SetResult(&result).
		Patch(fmt.Sprintf("/approval-rulesets/%s",
			url.PathEscape(viper.GetString("approval-ruleset-id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error creating approval ruleset proposal: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))
	cli.PrintSeparatorln(printer)
	cli.PrintCelebrationlnf(printer, "Approval ruleset proposal (ID=%v) created!",
		approvalRulesetProposalCreateCmd_getProposalID(result))

	return nil
}

func approvalRulesetProposalCreateCmd_createBody() map[string]interface{} {
	return map[string]interface{}{
		"version": map[string]interface{}{},
	}
}

func approvalRulesetProposalCreateCmd_getProposalID(resource map[string]interface{}) interface{} {
	version := resource["version"].(map[string]interface{})
	return version["id"]
}

func approvalRulesetProposalCreateCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"approval-ruleset-id"},
	})
}

func init() {
	cmd := approvalRulesetProposalCreateCmd
	flags := cmd.Flags()
	approvalRulesetProposalCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("approval-ruleset-id", "", "approval ruleset ID (required)")
}
