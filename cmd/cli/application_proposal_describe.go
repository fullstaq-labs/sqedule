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

// applicationProposalDescribeCmd represents the 'application proposal describe' command
var applicationProposalDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe an application proposal",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return applicationProposalDescribeCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationProposalDescribeCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := applicationProposalDescribeCmd_checkConfig(viper)
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
		Get(fmt.Sprintf("/applications/%s/proposals/%s",
			url.PathEscape(viper.GetString("application-id")),
			url.PathEscape(viper.GetString("id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error describing application proposal: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))

	return nil
}

func applicationProposalDescribeCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"application-id", "id"},
	})
}

func init() {
	cmd := applicationProposalDescribeCmd
	flags := cmd.Flags()
	applicationProposalCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("application-id", "", "application ID (required)")
	flags.String("id", "", "proposal ID (required)")
}
