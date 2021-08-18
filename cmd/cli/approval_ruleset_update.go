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

// approvalRulesetUpdateCmd represents the 'approval-ruleset update' command
var approvalRulesetUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an approval ruleset's non-versioned properties",
	Long: `Update an approval ruleset's non-versioned properties.

To update its versioned properties (e.g. display name or rules):

 1. create a proposal first: ` + "`sqedule approval-ruleset proposal create`" + `
 2. then use ` + "`sqedule approval-ruleset proposal update`" + ` and ` + "`sqedule approval-ruleset proposal rule`",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return approvalRulesetUpdateCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func approvalRulesetUpdateCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := approvalRulesetUpdateCmd_checkConfig(viper)
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
		SetBody(approvalRulesetUpdateCmd_createBody(viper)).
		SetResult(&result).
		Patch(fmt.Sprintf("/approval-rulesets/%s",
			url.PathEscape(viper.GetString("id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error updating approval ruleset: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))
	cli.PrintSeparatorln(printer)
	cli.PrintCelebrationlnf(printer, "Approval ruleset '%s' updated!", viper.GetString("id"))

	return nil
}

func approvalRulesetUpdateCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"id"},
	})
}

func approvalRulesetUpdateCmd_createBody(viper *viper.Viper) json.ApprovalRulesetInput {
	return json.ApprovalRulesetInput{
		ID: cli.GetViperStringIfSet(viper, "set-id"),
	}
}

func init() {
	cmd := approvalRulesetUpdateCmd
	flags := cmd.Flags()
	approvalRulesetCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("id", "", "ID of approval ruleset to update (required)")
	flags.String("set-id", "", "change approval ruleset ID")
}
