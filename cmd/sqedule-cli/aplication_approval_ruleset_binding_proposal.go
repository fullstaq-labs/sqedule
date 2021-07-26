package main

import (
	"fmt"
	"net/url"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applicationApprovalRulesetBindingProposalCmd represents the 'application-approval-ruleset-binding proposal' command
var applicationApprovalRulesetBindingProposalCmd = &cobra.Command{
	Use:   "proposal",
	Short: "Manage proposals",
}

func applicationApprovalRulesetBindingProposalCmd_getRuleset(viper *viper.Viper, config cli.Config, state cli.State) (map[string]interface{}, error) {
	req, err := cli.NewApiRequest(config, state)
	if err != nil {
		return nil, err
	}

	var ruleset map[string]interface{}
	resp, err := req.
		SetResult(&ruleset).
		Get(fmt.Sprintf("/application-approval-ruleset-bindings/%s/proposals/%s",
			url.PathEscape(viper.GetString("application-approval-ruleset-binding-id")),
			url.PathEscape(viper.GetString("proposal-id"))))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("Error reading application approval ruleset binding: %s", cli.GetApiErrorMessage(resp))
	}

	return ruleset, nil
}

func init() {
	applicationApprovalRulesetBindingCmd.AddCommand(applicationApprovalRulesetBindingProposalCmd)
}
