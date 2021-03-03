package httpapi

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels"
)

type releaseJSON struct {
	Application    *applicationJSON `json:"application,omitempty"`
	ID             uint64           `json:"id"`
	State          string           `json:"state"`
	SourceIdentity *string          `json:"source_identity"`
	Comments       *string          `json:"comments"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	FinalizedAt    *time.Time       `json:"finalized_at"`
}

func createReleaseJSONFromDbModel(release dbmodels.Release, includeApplication bool) releaseJSON {
	result := releaseJSON{
		ID:        release.ID,
		State:     string(release.State),
		CreatedAt: release.CreatedAt,
		UpdatedAt: release.UpdatedAt,
	}
	if includeApplication {
		applicationJSON := createApplicationJSONFromDbModel(release.Application)
		result.Application = &applicationJSON
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

func patchReleaseDbModelFromJSON(release *dbmodels.Release, json releaseJSON) {
	if json.SourceIdentity != nil {
		release.SourceIdentity = sql.NullString{String: *json.SourceIdentity, Valid: true}
	}
	if json.Comments != nil {
		release.Comments = sql.NullString{String: *json.Comments, Valid: true}
	}
}
