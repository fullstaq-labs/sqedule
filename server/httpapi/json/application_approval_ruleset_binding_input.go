package json

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
)

//
// ******** Types, constants & variables ********
//

type ApplicationApprovalRulesetBindingInput struct {
	Version *ApplicationApprovalRulesetBindingVersionInput `json:"version"`
}

type ApplicationApprovalRulesetBindingVersionInput struct {
	ReviewableVersionInputBase
	Mode    *approvalrulesetbindingmode.Mode `json:"mode"`
	Enabled *bool                            `json:"enabled"`
}

//
// ******** Other functions ********
//

func PatchApplicationApprovalRulesetBinding(binding *dbmodels.ApplicationApprovalRulesetBinding, input ApplicationApprovalRulesetBindingInput) {
	// Nothing to do
}

func PatchApplicationApprovalRulesetBindingAdjustment(organizationID string, adjustment *dbmodels.ApplicationApprovalRulesetBindingAdjustment, input ApplicationApprovalRulesetBindingVersionInput) {
	if input.Mode != nil {
		adjustment.Mode = *input.Mode
	}
	if input.Enabled != nil {
		adjustment.Enabled = input.Enabled
	}
}
