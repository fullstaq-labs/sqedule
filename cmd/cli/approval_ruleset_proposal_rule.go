package main

import (
	"fmt"
	"net/url"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// approvalRulesetProposalRuleCmd represents the 'approval-ruleset proposal rule' command
var approvalRulesetProposalRuleCmd = &cobra.Command{
	Use:   "rule",
	Short: "Manage rules in an approval ruleset proposal",
}

func approvalRulesetProposalRuleCmd_getRules(ruleset map[string]interface{}) ([]map[string]interface{}, error) {
	version := ruleset["version"].(map[string]interface{})
	rules := version["approval_rules"].([]interface{})
	result := make([]map[string]interface{}, 0, len(rules))
	for _, rule := range rules {
		result = append(result, rule.(map[string]interface{}))
	}
	return result, nil
}

func approvalRulesetProposalRuleCmd_patchRules(viper *viper.Viper, config cli.Config, state cli.State, rules []map[string]interface{}) (map[string]interface{}, error) {
	req, err := cli.NewApiRequest(config, state)
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"approval_rules": rules,
	}

	var ruleset map[string]interface{}
	resp, err := req.
		SetBody(&body).
		SetResult(&ruleset).
		Patch(fmt.Sprintf("/approval-rulesets/%s/proposals/%s",
			url.PathEscape(viper.GetString("approval-ruleset-id")),
			url.PathEscape(viper.GetString("proposal-id"))))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("Error updating ruleset: %s", cli.GetApiErrorMessage(resp))
	}

	return ruleset, nil
}

func init() {
	approvalRulesetProposalCmd.AddCommand(approvalRulesetProposalRuleCmd)
}
