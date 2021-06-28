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

// approvalRulesetVersionDescribeCmd represents the 'approval-ruleset version describe' command
var approvalRulesetVersionDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe an approval ruleset version",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return approvalRulesetVersionDescribeCmd_run(viper.GetViper(), mocking.RealPrinter{}, false)
	},
}

func approvalRulesetVersionDescribeCmd_run(viper *viper.Viper, printer mocking.IPrinter, testing bool) error {
	err := approvalRulesetVersionDescribeCmd_checkConfig(viper)
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
		Get(fmt.Sprintf("/approval-rulesets/%s/versions/%d",
			url.PathEscape(viper.GetString("approval-ruleset-id")),
			viper.GetUint("version-number")))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error describing approval ruleset version: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.Println(string(output))

	return nil
}

func approvalRulesetVersionDescribeCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"approval-ruleset-id"},
		UintNonZero:    []string{"version-number"},
	})
}

func init() {
	cmd := approvalRulesetVersionDescribeCmd
	flags := cmd.Flags()
	approvalRulesetVersionCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("approval-ruleset-id", "", "")
	flags.Uint("version-number", 0, "")
}
