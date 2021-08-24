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

// applicationProposalReviewCmd represents the 'application proposal review' command
var applicationProposalReviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review an application proposal",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return applicationProposalReviewCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationProposalReviewCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := applicationProposalReviewCmd_checkConfig(viper)
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
		SetBody(applicationProposalReviewCmd_createBody(viper)).
		SetResult(&result).
		Put(fmt.Sprintf("/applications/%s/proposals/%s/state",
			url.PathEscape(viper.GetString("application-id")),
			url.PathEscape(viper.GetString("id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error reviewing application proposal: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))
	cli.PrintSeparatorln(printer)
	cli.PrintCelebrationlnf(printer, "Application proposal reviewed!")

	return nil
}

func applicationProposalReviewCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"application-id", "id", "action"},
	})
}

func applicationProposalReviewCmd_createBody(viper *viper.Viper) json.ReviewableProposalStateInput {
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
	cmd := applicationProposalReviewCmd
	flags := cmd.Flags()
	applicationProposalCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("application-id", "", "application ID (required)")
	flags.String("id", "", "proposal ID (required)")
	flags.String("action", "", "'approve' or 'reject'")
}
