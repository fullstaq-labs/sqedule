package main

import (
	"fmt"
	"time"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// loginCmd represents the 'login' command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Logs into a Sqedule server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return loginCmd_run(cmd, args, viper.GetViper())
	},
}

func loginCmd_run(cmd *cobra.Command, args []string, viper *viper.Viper) error {
	config := cli.LoadConfigFromViper(viper)
	state, err := cli.LoadStateFromFilesystem()
	if err != nil {
		return fmt.Errorf("Error loading state: %w", err)
	}

	err = loginCmd_checkConfig(viper, config)
	if err != nil {
		return err
	}

	req, err := cli.NewApiRequest(config, state)
	if err != nil {
		return err
	}

	var result struct {
		Token  string
		Expire string
	}

	resp, err := req.
		SetBody(loginCmd_createBody(viper)).
		SetResult(&result).
		Post("/auth/login")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error logging in: %s", cli.GetApiErrorMessage(resp))
	}

	state.AuthToken = result.Token
	state.AuthTokenExpirationTime, err = time.Parse(time.RFC3339, result.Expire)
	if err != nil {
		return fmt.Errorf("Error parsing authentication token expiration time: %w", err)
	}

	err = state.SaveToFilesystem()
	if err != nil {
		return fmt.Errorf("Error saving state: %w", err)
	}

	return nil
}

func loginCmd_checkConfig(viper *viper.Viper, config cli.Config) error {
	err := cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"organization-id", "password"},
	})
	if err != nil {
		return err
	}

	err = cli.RequireOneOfConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"email", "service-account-name"},
	})
	if err != nil {
		return err
	}

	return nil
}

func loginCmd_createBody(viper *viper.Viper) interface{} {
	if len(viper.GetString("service-account-name")) > 0 {
		return map[string]interface{}{
			"organization_id":      viper.GetString("organization-id"),
			"service_account_name": viper.GetString("service-account-name"),
			"password":             viper.GetString("password"),
		}
	} else {
		return map[string]interface{}{
			"organization_id": viper.GetString("organization-id"),
			"email":           viper.GetString("email"),
			"password":        viper.GetString("password"),
		}
	}
}

func init() {
	cmd := loginCmd
	flags := cmd.Flags()
	rootCmd.AddCommand(cmd)

	cli.DefineServerFlags(cmd)

	flags.String("organization-id", "", "organization ID (required)")
	flags.StringP("email", "e", "", "user account email")
	flags.StringP("service-account-name", "n", "", "service account name")
	flags.StringP("password", "p", "", "user or service account password (required)")

	viper.BindPFlags(flags)
}
