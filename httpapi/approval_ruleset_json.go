package httpapi

import (
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels"
)

type approvalRulesetJSON struct {
	ID                 string    `json:"id"`
	MajorVersionNumber *uint32   `json:"major_version_number"`
	MinorVersionNumber uint32    `json:"minor_version_number"`
	DisplayName        string    `json:"display_name"`
	Description        string    `json:"description"`
	GloballyApplicable bool      `json:"globally_applicable"`
	ReviewState        string    `json:"review_state"`
	ReviewComments     *string   `json:"review_comments"`
	Enabled            bool      `json:"enabled"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func createApprovalRulesetJSONFromDbModel(ruleset dbmodels.ApprovalRuleset, majorVersion dbmodels.ApprovalRulesetMajorVersion, minorVersion dbmodels.ApprovalRulesetMinorVersion) approvalRulesetJSON {
	result := approvalRulesetJSON{
		ID:                 ruleset.ID,
		MajorVersionNumber: majorVersion.VersionNumber,
		MinorVersionNumber: minorVersion.VersionNumber,
		DisplayName:        minorVersion.DisplayName,
		Description:        minorVersion.Description,
		GloballyApplicable: minorVersion.GloballyApplicable,
		ReviewState:        string(minorVersion.ReviewState),
		CreatedAt:          ruleset.CreatedAt,
		UpdatedAt:          minorVersion.CreatedAt,
	}
	return result
}
