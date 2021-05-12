package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// releaseWaitCmd represents the 'release wait' command
var releaseWaitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Wait until a release's approval state is final",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		state, err := releaseWaitCmd_run(viper.GetViper(), mocking.RealPrinter{}, mocking.RealClock{}, false)
		if err != nil {
			return err
		}

		if viper.GetBool("fail-if-unapproved") && state != releasestate.Approved {
			os.Exit(40)
		}

		return nil
	},
}

func releaseWaitCmd_run(viper *viper.Viper, printer mocking.IPrinter, clock mocking.IClock,
	invokedByCreate bool) (releasestate.State, error) {

	err := releaseWaitCmd_checkConfig(viper, false)
	if err != nil {
		return releasestate.InProgress, err
	}

	config := cli.LoadConfigFromViper(viper)
	state, err := cli.LoadStateFromFilesystem()
	if err != nil {
		return releasestate.InProgress, fmt.Errorf("Error loading state: %w", err)
	}

	req, err := cli.NewApiRequest(config, state)
	if err != nil {
		return releasestate.InProgress, err
	}

	if invokedByCreate {
		releaseWaitCmd_sleep(viper, clock)
	}

	deadline, err := releaseWaitCmd_getDeadline(viper, clock)
	if err != nil {
		return releasestate.InProgress, err
	}

	for {
		var release json.ReleaseWithAssociations
		resp, err := req.
			SetResult(&release).
			Get(fmt.Sprintf("/applications/%s/releases/%s",
				url.PathEscape(viper.GetString("application-id")),
				url.PathEscape(strconv.FormatUint(uint64(viper.GetUint("release-id")), 10))))
		if err != nil {
			return releasestate.InProgress, err
		}
		if resp.IsError() {
			return releasestate.InProgress, fmt.Errorf("Error querying release: %s", cli.GetApiErrorMessage(resp))
		}

		printer.Printf("Current state: %v\n", release.State)
		if release.ApprovalStatusIsFinal() {
			return releasestate.State(release.State), nil
		} else if (deadline != time.Time{}) && time.Now().After(deadline) {
			return releasestate.InProgress, errors.New("Timeout")
		} else {
			releaseWaitCmd_sleep(viper, clock)
		}
	}
}

func releaseWaitCmd_checkConfig(viper *viper.Viper, fromCreateCmd bool) error {
	var uintNonZeroOptions []string
	if !fromCreateCmd {
		uintNonZeroOptions = append(uintNonZeroOptions, "release-id")
	}
	err := cli.RequireConfigOptions(viper, cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"application-id"},
		UintNonZero:    uintNonZeroOptions,
	})
	if err != nil {
		return err
	}

	_, err = releaseWaitCmd_getTimeout(viper)
	if err != nil {
		return err
	}

	_, err = releaseWaitCmd_getInterval(viper)
	if err != nil {
		return err
	}

	return nil
}

func releaseWaitCmd_getTimeout(viper *viper.Viper) (time.Duration, error) {
	timeout := viper.GetString("wait-timeout")
	result, err := time.ParseDuration(timeout)
	if err != nil {
		return time.Duration(0), fmt.Errorf("Error parsing configuration wait-timeout: %w", err)
	}
	return result, nil
}

func releaseWaitCmd_getInterval(viper *viper.Viper) (time.Duration, error) {
	interval := viper.GetString("wait-interval")
	result, err := time.ParseDuration(interval)
	if err != nil {
		return time.Duration(0), fmt.Errorf("Error parsing configuration wait-interval: %w", err)
	}
	if result == time.Duration(0) {
		return time.Duration(0), errors.New("Configuration wait-interval must be larger than 0")
	}
	return result, nil
}

func releaseWaitCmd_getDeadline(viper *viper.Viper, clock mocking.IClock) (time.Time, error) {
	timeout, err := releaseWaitCmd_getTimeout(viper)
	if err != nil {
		return time.Time{}, err
	}

	if timeout == time.Duration(0) {
		return time.Time{}, nil
	} else {
		return clock.Now().Add(timeout), nil
	}
}

func releaseWaitCmd_sleep(viper *viper.Viper, clock mocking.IClock) {
	interval, err := releaseWaitCmd_getInterval(viper)
	if err != nil {
		panic("Bug: cannot parse wait-interval")
	}
	clock.Sleep(interval)
}

func releaseWaitCmd_defineFlagsSharedWithCreateCmd(flags *pflag.FlagSet) {
	flags.String("wait-timeout", "0", "Max time to wait until timeout, 0 means forever. Non-zero values must contain a time unit (e.g. 1s, 2m, 3h)")
	flags.String("wait-interval", "1m", "Time between state checks. Must contain a time unit (e.g. 1s, 2m, 3h)")
}

func init() {
	cmd := releaseWaitCmd
	flags := cmd.Flags()
	releaseCmd.AddCommand(cmd)

	cli.DefineServerFlags(flags)

	flags.StringP("application-id", "a", "", "ID of application in which the release is located (required)")
	flags.Uint("release-id", 0, "ID of release to wait for (required)")
	flags.Bool("fail-if-unapproved", false, "exit with code 40 if finalized release is not in the 'approved' state")

	releaseWaitCmd_defineFlagsSharedWithCreateCmd(flags)
}
