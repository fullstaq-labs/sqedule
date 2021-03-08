package httpapi

import (
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels"
)

type applicationJSON struct {
	ID                      string                                   `json:"id"`
	MajorVersionNumber      *uint32                                  `json:"major_version_number"`
	MinorVersionNumber      uint32                                   `json:"minor_version_number"`
	DisplayName             *string                                  `json:"display_name"`
	Enabled                 *bool                                    `json:"enabled"`
	CreatedAt               time.Time                                `json:"created_at"`
	UpdatedAt               time.Time                                `json:"updated_at"`
	ApprovalRulesetBindings *[]applicationApprovalRulesetBindingJSON `json:"approval_ruleset_bindings,omitempty"`
}

func createApplicationJSONFromDbModel(application dbmodels.Application, majorVersion dbmodels.ApplicationMajorVersion, minorVersion dbmodels.ApplicationMinorVersion,
	rulesetBindings *[]dbmodels.ApplicationApprovalRulesetBinding) applicationJSON {

	result := applicationJSON{
		ID:                 application.ID,
		MajorVersionNumber: majorVersion.VersionNumber,
		MinorVersionNumber: minorVersion.VersionNumber,
		DisplayName:        &minorVersion.DisplayName,
		Enabled:            &minorVersion.Enabled,
		CreatedAt:          application.CreatedAt,
		UpdatedAt:          minorVersion.CreatedAt,
	}
	if rulesetBindings != nil {
		rulesetBindingsJSON := make([]applicationApprovalRulesetBindingJSON, 0, len(*rulesetBindings))
		for _, rulesetBinding := range *rulesetBindings {
			rulesetBindingsJSON = append(rulesetBindingsJSON,
				createApplicationApprovalRulesetBindingJSONFromDbModel(rulesetBinding,
					*rulesetBinding.LatestMajorVersion, *rulesetBinding.LatestMinorVersion, false))
		}
		result.ApprovalRulesetBindings = &rulesetBindingsJSON
	}
	return result
}
