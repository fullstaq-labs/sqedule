package main

import (
	encjson "encoding/json"
	"fmt"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json/proposalstate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// approvalRulesetCreateCmd represents the 'approval-ruleset create' command
var approvalRulesetCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an approval ruleset",
	Long:  "Creates an approval ruleset without any rules inside it. To add rules, use `sqedule approval-ruleset proposal rule create`",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return approvalRulesetCreateCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func approvalRulesetCreateCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := approvalRulesetCreateCmd_checkConfig(viper)
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
		SetBody(approvalRulesetCreateCmd_createBody(viper)).
		SetResult(&ruleset).
		Post("/approval-rulesets")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error creating approval ruleset: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(ruleset, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))
	cli.PrintSeparatorln(printer)
	cli.PrintCelebrationlnf(printer, "Approval ruleset '%s' created!", viper.GetString("id"))

	version := ruleset["version"].(map[string]interface{})
	if version["version_state"] == "approved" {
		cli.PrintTiplnf(printer, "It has been auto-approved by the system. To view it, use `sqedule approval-ruleset describe`")
	} else {
		cli.PrintTiplnf(printer, "It is still a proposal. To view it, use `sqedule approval-ruleset proposal list`")
	}
	cli.PrintCaveatlnf(printer, "It has no rules yet. To add rules, use `sqedule approval-ruleset proposal rule create-...`")

	return nil
}

func approvalRulesetCreateCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"id", "display-name"},
	})
}

func approvalRulesetCreateCmd_createBody(viper *viper.Viper) json.ApprovalRulesetInput {
	return json.ApprovalRulesetInput{
		ID: lib.NewStringPtr(viper.GetString("id")),
		Version: &json.ApprovalRulesetVersionInput{
			ReviewableVersionInputBase: json.ReviewableVersionInputBase{
				ProposalState: proposalstate.State(viper.GetString("proposal-state")),
			},
			DisplayName: lib.NewStringPtr(viper.GetString("display-name")),
			Enabled:     lib.NewBoolPtr(viper.GetBool("enabled")),
		},
	}
}

func init() {
	cmd := approvalRulesetCreateCmd
	flags := cmd.Flags()
	approvalRulesetCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("id", "", "A machine-friendly identifier (required)")
	flags.String("display-name", "", "A human-friendly display name (required)")
	flags.String("description", "", "")
	flags.String("proposal-state", "draft", "'draft', 'final' or 'abandon'")
	flags.Bool("enabled", true, "Whether to enable this ruleset")
}
