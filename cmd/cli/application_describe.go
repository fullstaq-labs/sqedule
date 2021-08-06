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

// applicationDescribeCmd represents the 'application describe' command
var applicationDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe an aplication",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		return applicationDescribeCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationDescribeCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := applicationDescribeCmd_checkConfig(viper)
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
		Get(fmt.Sprintf("/applications/%s",
			url.PathEscape(viper.GetString("id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error describing application: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))

	return nil
}

func applicationDescribeCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"id"},
	})
}

func init() {
	cmd := applicationDescribeCmd
	flags := cmd.Flags()
	applicationCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("id", "", "Application ID (required)")
}
