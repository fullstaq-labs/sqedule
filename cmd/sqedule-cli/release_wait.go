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

	var lastSleepDuration time.Duration
	if invokedByCreate {
		lastSleepDuration = releaseWaitCmd_sleep(viper, clock, time.Duration(0))
	}

	deadline, err := releaseWaitCmd_getDeadline(viper, clock)
	if err != nil {
		return releasestate.InProgress, err
	}

	for {
		req, err := cli.NewApiRequest(config, state)
		if err != nil {
			return releasestate.InProgress, err
		}

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

		printer.PrintMessagef("Current state: %v\n", release.State)
		if release.ApprovalStatusIsFinal() {
			return releasestate.State(release.State), nil
		} else if (deadline != time.Time{}) && time.Now().After(deadline) {
			return releasestate.InProgress, errors.New("Timeout")
		} else {
			lastSleepDuration = releaseWaitCmd_sleep(viper, clock, lastSleepDuration)
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

	_, err = releaseWaitCmd_getMinDuration(viper)
	if err != nil {
		return err
	}

	_, err = releaseWaitCmd_getMaxDuration(viper)
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

func releaseWaitCmd_getMinDuration(viper *viper.Viper) (time.Duration, error) {
	interval := viper.GetString("wait-min-duration")
	result, err := time.ParseDuration(interval)
	if err != nil {
		return time.Duration(0), fmt.Errorf("Error parsing configuration wait-min-duration: %w", err)
	}
	if result == time.Duration(0) {
		return time.Duration(0), errors.New("Configuration wait-min-duration must be larger than 0")
	}
	return result, nil
}

func releaseWaitCmd_getMaxDuration(viper *viper.Viper) (time.Duration, error) {
	interval := viper.GetString("wait-max-duration")
	result, err := time.ParseDuration(interval)
	if err != nil {
		return time.Duration(0), fmt.Errorf("Error parsing configuration wait-max-duration: %w", err)
	}
	if result == time.Duration(0) {
		return time.Duration(0), errors.New("Configuration wait-max-duration must be larger than 0")
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

func releaseWaitCmd_sleep(viper *viper.Viper, clock mocking.IClock, prevDuration time.Duration) time.Duration {
	var duration time.Duration
	var err error

	if prevDuration == time.Duration(0) {
		duration, err = releaseWaitCmd_getMinDuration(viper)
		if err != nil {
			panic("Bug: cannot parse wait-min-duration")
		}
	} else {
		duration = prevDuration * 2
		maxDuration, err := releaseWaitCmd_getMaxDuration(viper)
		if err != nil {
			panic("Bug: cannot parse wait-max-duration")
		}
		if duration > maxDuration {
			duration = maxDuration
		}
	}
	clock.Sleep(duration)
	return duration
}

func releaseWaitCmd_defineFlagsSharedWithCreateCmd(flags *pflag.FlagSet) {
	flags.String("wait-timeout", "0", "Max time to wait until timeout, 0 means forever. Non-zero values must contain a time unit (e.g. 1s, 2m, 3h)")
	flags.String("wait-min-duration", "1s", "Minimum duration between state checks. Must contain a time unit (e.g. 1s, 2m, 3h)")
	flags.String("wait-max-duration", "1m", "Maximum duration between state checks. Must contain a time unit (e.g. 1s, 2m, 3h)")
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
