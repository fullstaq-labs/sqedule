package json

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
)

//
// ******** Types, constants & variables ********
//

type Release struct {
	ID    uint64 `json:"id"`
	State string `json:"state"`
	ReleasePatchablePart
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	FinalizedAt *time.Time `json:"finalized_at"`
}

type ReleasePatchablePart struct {
	SourceIdentity *string                 `json:"source_identity"`
	Metadata       *map[string]interface{} `json:"metadata"`
	Comments       *string                 `json:"comments"`
}

type ReleaseWithApplicationAssociation struct {
	Release
	Application ApplicationWithLatestApprovedVersion `json:"application"`
}

type ReleaseWithAssociations struct {
	Release
	Application             *ApplicationWithLatestApprovedVersion                  `json:"application,omitempty"`
	ApprovalRulesetBindings *[]ReleaseApprovalRulesetBindingWithRulesetAssociation `json:"approval_ruleset_bindings,omitempty"`
}

//
// ******** Release methods ********
//

func (release Release) ApprovalStatusIsFinal() bool {
	return releasestate.State(release.State).IsFinal()
}

//
// ******** Constructor functions ********
//

func CreateFromDbRelease(release dbmodels.Release) Release {
	result := Release{
		ID:        release.ID,
		State:     string(release.State),
		CreatedAt: release.CreatedAt,
		UpdatedAt: release.UpdatedAt,
	}
	if release.Metadata != nil {
		result.Metadata = (*map[string]interface{})(&release.Metadata)
	} else {
		emptyMap := map[string]interface{}{}
		result.Metadata = &emptyMap
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
	if release.Application.Version == nil {
		panic("Associated application must have an associated version")
	}
	if release.Application.Version.VersionNumber == nil {
		panic("Associated application's associated version must be finalized")
	}
	if release.Application.Version.Adjustment == nil {
		panic("Associated application must have an associated adjustment")
	}

	return ReleaseWithApplicationAssociation{
		Release:     CreateFromDbRelease(release),
		Application: CreateApplicationWithLatestApprovedVersion(release.Application, release.Application.Version),
	}
}

func CreateFromDbReleaseWithAssociations(release dbmodels.Release, includeApplication bool,
	rulesetBindings *[]dbmodels.ReleaseApprovalRulesetBinding) ReleaseWithAssociations {

	result := ReleaseWithAssociations{
		Release: CreateFromDbRelease(release),
	}
	if includeApplication {
		if release.Application.Version == nil {
			panic("Associated application must have an associated version")
		}
		if release.Application.Version.VersionNumber == nil {
			panic("Associated application's associated version must be finalized")
		}
		if release.Application.Version.Adjustment == nil {
			panic("Associated application must have an associated adjustment")
		}
		applicationJSON := CreateApplicationWithLatestApprovedVersion(release.Application, release.Application.Version)
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

//
// ******** Other functions ********
//

func PatchDbRelease(release *dbmodels.Release, json ReleasePatchablePart) {
	if json.SourceIdentity != nil {
		release.SourceIdentity = sql.NullString{String: *json.SourceIdentity, Valid: true}
	}
	if json.Metadata != nil {
		release.Metadata = *json.Metadata
	}
	if json.Comments != nil {
		release.Comments = sql.NullString{String: *json.Comments, Valid: true}
	}
}
