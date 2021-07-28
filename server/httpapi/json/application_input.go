package json

import (
	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

//
// ******** Types, constants & variables ********
//

type ApplicationInput struct {
	ID      *string                  `json:"id"`
	Version *ApplicationVersionInput `json:"version"`
}

type ApplicationVersionInput struct {
	ReviewableVersionInputBase
	DisplayName *string `json:"display_name"`
	Enabled     *bool   `json:"enabled"`
}

//
// ******** Other functions ********
//

func PatchApplication(app *dbmodels.Application, input ApplicationInput) {
	if input.ID != nil {
		app.ID = *input.ID
	}
}

func PatchApplicationAdjustment(organizationID string, adjustment *dbmodels.ApplicationAdjustment, input ApplicationVersionInput) {
	if input.DisplayName != nil {
		adjustment.DisplayName = *input.DisplayName
	}
	if input.Enabled != nil {
		adjustment.Enabled = lib.CopyBoolPtr(input.Enabled)
	}
}
