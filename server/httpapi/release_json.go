package httpapi

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

type releaseJSON struct {
	ID             uint64     `json:"id"`
	State          string     `json:"state"`
	SourceIdentity *string    `json:"source_identity"`
	Comments       *string    `json:"comments"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	FinalizedAt    *time.Time `json:"finalized_at"`
}

type releaseWithApplicationAssociationJSON struct {
	releaseJSON
	Application applicationJSON `json:"application"`
}

type releaseWithAssociationsJSON struct {
	releaseJSON
	Application             *applicationJSON                                           `json:"application,omitempty"`
	ApprovalRulesetBindings *[]releaseApprovalRulesetBindingWithRulesetAssociationJSON `json:"approval_ruleset_bindings,omitempty"`
}

func createReleaseJSONFromDbModel(release dbmodels.Release) releaseJSON {
	result := releaseJSON{
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

func createReleaseWithApplicationAssociationJSONFromDbModel(release dbmodels.Release) releaseWithApplicationAssociationJSON {

	if release.Application.LatestMajorVersion == nil {
		panic("Associated application must have an associated latest major version")
	}
	if release.Application.LatestMajorVersion.VersionNumber == nil {
		panic("Associated application's associated latest major version must be finalized")
	}
	if release.Application.LatestMinorVersion == nil {
		panic("Associated application must have an associated latest minor version")
	}

	return releaseWithApplicationAssociationJSON{
		releaseJSON: createReleaseJSONFromDbModel(release),
		Application: createApplicationJSONFromDbModel(release.Application,
			*release.Application.LatestMajorVersion, *release.Application.LatestMinorVersion,
			nil),
	}
}

func createReleaseWithAssociationsJSONFromDbModel(release dbmodels.Release, includeApplication bool,
	rulesetBindings *[]dbmodels.ReleaseApprovalRulesetBinding) releaseWithAssociationsJSON {

	result := releaseWithAssociationsJSON{
		releaseJSON: createReleaseJSONFromDbModel(release),
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
		applicationJSON := createApplicationJSONFromDbModel(release.Application,
			*release.Application.LatestMajorVersion, *release.Application.LatestMinorVersion,
			nil)
		result.Application = &applicationJSON
	}
	if rulesetBindings != nil {
		rulesetBindingsJSON := make([]releaseApprovalRulesetBindingWithRulesetAssociationJSON, 0, len(*rulesetBindings))
		for _, rulesetBinding := range *rulesetBindings {
			rulesetBindingsJSON = append(rulesetBindingsJSON,
				createReleaseApprovalRulesetBindingWithRulesetAssociationJSONFromDbModel(rulesetBinding))
		}
		result.ApprovalRulesetBindings = &rulesetBindingsJSON
	}
	return result
}

func patchReleaseDbModelFromJSON(release *dbmodels.Release, json releaseJSON) {
	if json.SourceIdentity != nil {
		release.SourceIdentity = sql.NullString{String: *json.SourceIdentity, Valid: true}
	}
	if json.Comments != nil {
		release.Comments = sql.NullString{String: *json.Comments, Valid: true}
	}
}
