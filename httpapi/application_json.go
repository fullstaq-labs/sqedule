package httpapi

import (
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels"
)

type applicationJSON struct {
	ID                       string    `json:"id"`
	LatestMajorVersionNumber uint32    `json:"latest_major_version_number"`
	LatestMinorVersionNumber uint32    `json:"latest_minor_version_number"`
	DisplayName              *string   `json:"display_name"`
	Enabled                  *bool     `json:"enabled"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

func createApplicationJSONFromDbModel(application dbmodels.Application) applicationJSON {
	if application.LatestMajorVersion == nil {
		panic("Given application must have an associated latest major version")
	}
	if application.LatestMajorVersion.VersionNumber == nil {
		panic("Given application's associated latest major version must be finalized")
	}
	if application.LatestMinorVersion == nil {
		panic("Given application must have an associated latest minor version")
	}
	result := applicationJSON{
		ID:                       application.ID,
		LatestMajorVersionNumber: *application.LatestMajorVersion.VersionNumber,
		LatestMinorVersionNumber: application.LatestMinorVersion.VersionNumber,
		DisplayName:              &application.LatestMinorVersion.DisplayName,
		Enabled:                  &application.LatestMinorVersion.Enabled,
		CreatedAt:                application.CreatedAt,
		UpdatedAt:                application.LatestMinorVersion.CreatedAt,
	}
	return result
}
