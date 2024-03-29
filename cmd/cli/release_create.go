package main

import (
	encjson "encoding/json"
	"fmt"
	"net/url"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// releaseCreateCmd represents the 'release create' command
var releaseCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a release",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			return err
		}

		return releaseCreateCmd_run(viper.GetViper(), mocking.RealPrinter{}, false)
	},
}

func releaseCreateCmd_run(viper *viper.Viper, printer mocking.IPrinter, testing bool) error {
	err := releaseCreateCmd_checkConfig(viper)
	if err != nil {
		return err
	}
	if viper.GetBool("wait") {
		err = releaseWaitCmd_checkConfig(viper, true)
		if err != nil {
			return err
		}
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

	body, err := releaseCreateCmd_createBody(viper)
	if err != nil {
		return err
	}

	var release map[string]interface{}
	resp, err := req.
		SetBody(body).
		SetResult(&release).
		Post(fmt.Sprintf("/applications/%s/releases",
			url.PathEscape(viper.GetString("application-id"))))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error creating release: %s", cli.GetApiErrorMessage(resp))
	}

	output, err := encjson.MarshalIndent(release, "", "    ")
	if err != nil {
		return fmt.Errorf("Error formatting result as JSON: %w", err)
	}
	printer.PrintOutputln(string(output))

	if viper.GetBool("wait") && !releasestate.State(release["state"].(string)).IsFinal() {
		printer.PrintMessageln("Waiting for the release's approval state to become final...")
		viper.Set("release-id", release["id"])
		if testing {
			return nil
		}
		_, err = releaseWaitCmd_run(viper, printer, mocking.RealClock{}, true)
		return err
	}

	return nil
}

func releaseCreateCmd_checkConfig(viper *viper.Viper) error {
	return cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"application-id"},
	})
}

func releaseCreateCmd_createBody(viper *viper.Viper) (json.ReleasePatchablePart, error) {
	result := json.ReleasePatchablePart{
		SourceIdentity: lib.NonEmptyStringOrNil(viper.GetString("source-identity")),
		Comments:       lib.NonEmptyStringOrNil(viper.GetString("comments")),
	}
	if metadataText := viper.GetString("metadata"); len(metadataText) > 0 {
		var metadata map[string]interface{}
		err := encjson.Unmarshal([]byte(metadataText), &metadata)
		if err != nil {
			return json.ReleasePatchablePart{}, fmt.Errorf("Error parsing metadata as JSON object: %w", err)
		}
		result.Metadata = &metadata
	}
	return result, nil
}

func init() {
	cmd := releaseCreateCmd
	flags := cmd.Flags()
	releaseCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.StringP("application-id", "a", "", "ID of application for which to create a release (required)")
	flags.String("source-identity", "", "Source identity")
	flags.String("metadata", "", "Metadata (JSON object)")
	flags.String("comments", "", "Comments to add to the release")
	flags.BoolP("wait", "w", false, "Wait until the release's approval state is final")
	releaseWaitCmd_defineFlagsSharedWithCreateCmd(flags)
}
