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

// applicationVersionDescribeCmd represents the 'application version describe' command
var applicationVersionDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe an application version",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			return err
		}

		return applicationVersionDescribeCmd_run(viper.GetViper(), mocking.RealPrinter{})
	},
}

func applicationVersionDescribeCmd_run(viper *viper.Viper, printer mocking.IPrinter) error {
	err := applicationVersionDescribeCmd_checkConfig(viper)
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
		Get(fmt.Sprintf("/applications/%s/versions/%d",
			url.PathEscape(viper.GetString("application-id")),
			viper.GetUint("version-number")))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error describing application version: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))

	return nil
}

func applicationVersionDescribeCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"application-id"},
		UintNonZero:    []string{"version-number"},
	})
}

func init() {
	cmd := applicationVersionDescribeCmd
	flags := cmd.Flags()
	applicationVersionCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.String("application-id", "", "application ID (required)")
	flags.Uint("version-number", 0, "(required)")
}
