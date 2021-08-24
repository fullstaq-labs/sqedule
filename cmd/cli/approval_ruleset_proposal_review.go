package main

import (
	encjson "encoding/json"
	"fmt"
	"net/url"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json/reviewstateinput"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// approvalRulesetProposalReviewCmd represents the 'approval-ruleset proposal review' command
var approvalRulesetProposalReviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review an approval ruleset proposal",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return approvalRulesetProposalReviewCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func approvalRulesetProposalReviewCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := approvalRulesetProposalReviewCmd_checkConfig(viper)
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
		SetBody(approvalRulesetProposalReviewCmd_createBody(viper)).
		SetResult(&result).
		Put(fmt.Sprintf("/approval-rulesets/%s/proposals/%s/state",
			url.PathEscape(viper.GetString("approval-ruleset-id")),
			url.PathEscape(viper.GetString("id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error reviewing approval ruleset proposal: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))
	cli.PrintSeparatorln(printer)
	cli.PrintCelebrationlnf(printer, "Approval ruleset proposal reviewed!")

	return nil
}

func approvalRulesetProposalReviewCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"approval-ruleset-id", "id", "action"},
	})
}

func approvalRulesetProposalReviewCmd_createBody(viper *viper.Viper) json.ReviewableProposalStateInput {
	var state reviewstateinput.Input

	switch viper.GetString("action") {
	case "approve":
		state = reviewstateinput.Approved
	case "reject":
		state = reviewstateinput.Rejected
	default:
		panic("Unsupported action parameter '" + viper.GetString("action") + "'")
	}

	return json.ReviewableProposalStateInput{
		State: state,
	}
}

func init() {
	cmd := approvalRulesetProposalReviewCmd
	flags := cmd.Flags()
	approvalRulesetProposalCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("approval-ruleset-id", "", "approval ruleset ID (required)")
	flags.String("id", "", "proposal ID (required)")
	flags.String("action", "", "'approve' or 'reject'")
}
