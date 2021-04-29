package json

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

type Release struct {
	ID             uint64     `json:"id"`
	State          string     `json:"state"`
	SourceIdentity *string    `json:"source_identity"`
	Comments       *string    `json:"comments"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	FinalizedAt    *time.Time `json:"finalized_at"`
}

type ReleaseWithApplicationAssociation struct {
	Release
	Application Application `json:"application"`
}

type ReleaseWithAssociations struct {
	Release
	Application             *Application                                           `json:"application,omitempty"`
	ApprovalRulesetBindings *[]ReleaseApprovalRulesetBindingWithRulesetAssociation `json:"approval_ruleset_bindings,omitempty"`
}

func CreateFromDbRelease(release dbmodels.Release) Release {
	result := Release{
		ID:        release.ID,
		State:     string(release.State),
		CreatedAt: release.CreatedAt,
		UpdatedAt: release.UpdatedAt,
	}
	if release.SourceIdentity.Valid {
		result.SourceIdentity = &release.SourceIdentity.String
	}
	if release.Comments.Valid {
		result.Comments = &release.Comments.String
	}
	if release.FinalizedAt.Valid {
		result.FinalizedAt = &release.FinalizedAt.Time
	}
	return result
}

func CreateFromDbReleaseWithApplicationAssociation(release dbmodels.Release) ReleaseWithApplicationAssociation {
	if release.Application.LatestMajorVersion == nil {
		panic("Associated application must have an associated latest major version")
	}
	if release.Application.LatestMajorVersion.VersionNumber == nil {
		panic("Associated application's associated latest major version must be finalized")
	}
	if release.Application.LatestMinorVersion == nil {
		panic("Associated application must have an associated latest minor version")
	}

	return ReleaseWithApplicationAssociation{
		Release: CreateFromDbRelease(release),
		Application: CreateFromDbApplication(release.Application,
			*release.Application.LatestMajorVersion, *release.Application.LatestMinorVersion,
			nil),
	}
}

func CreateFromDbReleaseWithAssociations(release dbmodels.Release, includeApplication bool,
	rulesetBindings *[]dbmodels.ReleaseApprovalRulesetBinding) ReleaseWithAssociations {

	result := ReleaseWithAssociations{
		Release: CreateFromDbRelease(release),
	}
	if includeApplication {
		if release.Application.LatestMajorVersion == nil {
			panic("Associated application must have an associated latest major version")
		}
		if release.Application.LatestMajorVersion.VersionNumber == nil {
			panic("Associated application's associated latest major version must be finalized")
		}
		if release.Application.LatestMinorVersion == nil {
			panic("Associated application must have an associated latest minor version")
		}
		applicationJSON := CreateFromDbApplication(release.Application,
			*release.Application.LatestMajorVersion, *release.Application.LatestMinorVersion,
			nil)
		result.Application = &applicationJSON
	}
	if rulesetBindings != nil {
		rulesetBindingsJSON := make([]ReleaseApprovalRulesetBindingWithRulesetAssociation, 0, len(*rulesetBindings))
		for _, rulesetBinding := range *rulesetBindings {
			rulesetBindingsJSON = append(rulesetBindingsJSON,
				CreateFromDbReleaseApprovalRulesetBindingWithRulesetAssociation(rulesetBinding))
		}
		result.ApprovalRulesetBindings = &rulesetBindingsJSON
	}
	return result
}

func PatchDbRelease(release *dbmodels.Release, json Release) {
	if json.SourceIdentity != nil {
		release.SourceIdentity = sql.NullString{String: *json.SourceIdentity, Valid: true}
	}
	if json.Comments != nil {
		release.Comments = sql.NullString{String: *json.Comments, Valid: true}
	}
}
