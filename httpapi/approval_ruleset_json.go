package httpapi

import (
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels"
)

type approvalRulesetJSON struct {
	ID                       string    `json:"id"`
	LatestMajorVersionNumber uint32    `json:"latest_major_version_number"`
	LatestMinorVersionNumber uint32    `json:"latest_minor_version_number"`
	DisplayName              string    `json:"display_name"`
	Description              string    `json:"description"`
	GloballyApplicable       bool      `json:"globally_applicable"`
	ReviewState              string    `json:"review_state"`
	ReviewComments           *string   `json:"review_comments"`
	Enabled                  bool      `json:"enabled"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

func createApprovalRulesetJSONFromDbModel(ruleset dbmodels.ApprovalRuleset) approvalRulesetJSON {
	if ruleset.LatestMajorVersion == nil {
		panic("Given approval ruleset must have an associated latest major version")
	}
	if ruleset.LatestMajorVersion.VersionNumber == nil {
		panic("Given approval ruleset's associated latest major version must be finalized")
	}
	if ruleset.LatestMinorVersion == nil {
		panic("Given approval ruleset must have an associated latest minor version")
	}
	result := approvalRulesetJSON{
		ID:                       ruleset.ID,
		LatestMajorVersionNumber: *ruleset.LatestMajorVersion.VersionNumber,
		LatestMinorVersionNumber: ruleset.LatestMinorVersion.VersionNumber,
		DisplayName:              ruleset.LatestMinorVersion.DisplayName,
		Description:              ruleset.LatestMinorVersion.Description,
		GloballyApplicable:       ruleset.LatestMinorVersion.GloballyApplicable,
		ReviewState:              string(ruleset.LatestMinorVersion.ReviewState),
		CreatedAt:                ruleset.CreatedAt,
		UpdatedAt:                ruleset.LatestMinorVersion.CreatedAt,
	}
	return result
}
