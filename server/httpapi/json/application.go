package json

import (
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

type Application struct {
	ID                      string                                                     `json:"id"`
	VersionNumber           *uint32                                                    `json:"version_number"`
	AdjustmentNumber        uint32                                                     `json:"adjustment_number"`
	DisplayName             *string                                                    `json:"display_name"`
	Enabled                 *bool                                                      `json:"enabled"`
	ReviewState             string                                                     `json:"review_state"`
	ReviewComments          *string                                                    `json:"review_comments"`
	CreatedAt               time.Time                                                  `json:"created_at"`
	UpdatedAt               time.Time                                                  `json:"updated_at"`
	ApprovalRulesetBindings *[]ApplicationApprovalRulesetBindingWithRulesetAssociation `json:"approval_ruleset_bindings,omitempty"`
}

func CreateFromDbApplication(application dbmodels.Application, version dbmodels.ApplicationVersion, adjustment dbmodels.ApplicationAdjustment,
	rulesetBindings *[]dbmodels.ApplicationApprovalRulesetBinding) Application {

	var reviewComments *string
	if adjustment.ReviewComments.Valid {
		reviewComments = &adjustment.ReviewComments.String
	}

	result := Application{
		ID:               application.ID,
		VersionNumber:    version.VersionNumber,
		AdjustmentNumber: adjustment.AdjustmentNumber,
		DisplayName:      &adjustment.DisplayName,
		Enabled:          &adjustment.Enabled,
		ReviewState:      string(adjustment.ReviewState),
		ReviewComments:   reviewComments,
		CreatedAt:        application.CreatedAt,
		UpdatedAt:        adjustment.CreatedAt,
	}
	if rulesetBindings != nil {
		rulesetBindingsJSON := make([]ApplicationApprovalRulesetBindingWithRulesetAssociation, 0, len(*rulesetBindings))
		for _, rulesetBinding := range *rulesetBindings {
			rulesetBindingsJSON = append(rulesetBindingsJSON,
				CreateFromDbApplicationApprovalRulesetBindingWithRulesetAssociation(rulesetBinding,
					*rulesetBinding.LatestVersion, *rulesetBinding.LatestAdjustment))
		}
		result.ApprovalRulesetBindings = &rulesetBindingsJSON
	}
	return result
}
