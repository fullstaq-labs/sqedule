package main

import (
	"fmt"
	"net/url"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// approvalRulesetProposalCmd represents the 'approval-ruleset proposal' command
var approvalRulesetProposalCmd = &cobra.Command{
	Use:   "proposal",
	Short: "Manage proposals",
}

func approvalRulesetProposalCmd_getRuleset(viper *viper.Viper, config cli.Config, state cli.State) (map[string]interface{}, error) {
	req, err := cli.NewApiRequest(config, state)
	if err != nil {
		return nil, err
	}

	var ruleset map[string]interface{}
	resp, err := req.
		SetResult(&ruleset).
		Get(fmt.Sprintf("/approval-rulesets/%s/proposals/%s",
			url.PathEscape(viper.GetString("approval-ruleset-id")),
			url.PathEscape(viper.GetString("proposal-id"))))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("Error reading approval ruleset: %s", cli.GetApiErrorMessage(resp))
	}

	return ruleset, nil
}

func init() {
	approvalRulesetCmd.AddCommand(approvalRulesetProposalCmd)
}
