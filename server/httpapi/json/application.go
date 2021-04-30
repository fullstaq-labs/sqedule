package json

import (
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

type Application struct {
	ID                      string                                                     `json:"id"`
	MajorVersionNumber      *uint32                                                    `json:"major_version_number"`
	MinorVersionNumber      uint32                                                     `json:"minor_version_number"`
	DisplayName             *string                                                    `json:"display_name"`
	Enabled                 *bool                                                      `json:"enabled"`
	ReviewState             string                                                     `json:"review_state"`
	ReviewComments          *string                                                    `json:"review_comments"`
	CreatedAt               time.Time                                                  `json:"created_at"`
	UpdatedAt               time.Time                                                  `json:"updated_at"`
	ApprovalRulesetBindings *[]ApplicationApprovalRulesetBindingWithRulesetAssociation `json:"approval_ruleset_bindings,omitempty"`
}

func CreateFromDbApplication(application dbmodels.Application, majorVersion dbmodels.ApplicationMajorVersion, minorVersion dbmodels.ApplicationMinorVersion,
	rulesetBindings *[]dbmodels.ApplicationApprovalRulesetBinding) Application {

	var reviewComments *string
	if minorVersion.ReviewComments.Valid {
		reviewComments = &minorVersion.ReviewComments.String
	}

	result := Application{
		ID:                 application.ID,
		MajorVersionNumber: majorVersion.VersionNumber,
		MinorVersionNumber: minorVersion.VersionNumber,
		DisplayName:        &minorVersion.DisplayName,
		Enabled:            &minorVersion.Enabled,
		ReviewState:        string(minorVersion.ReviewState),
		ReviewComments:     reviewComments,
		CreatedAt:          application.CreatedAt,
		UpdatedAt:          minorVersion.CreatedAt,
	}
	if rulesetBindings != nil {
		rulesetBindingsJSON := make([]ApplicationApprovalRulesetBindingWithRulesetAssociation, 0, len(*rulesetBindings))
		for _, rulesetBinding := range *rulesetBindings {
			rulesetBindingsJSON = append(rulesetBindingsJSON,
				CreateFromDbApplicationApprovalRulesetBindingWithRulesetAssociation(rulesetBinding,
					*rulesetBinding.LatestMajorVersion, *rulesetBinding.LatestMinorVersion))
		}
		result.ApprovalRulesetBindings = &rulesetBindingsJSON
	}
	return result
}