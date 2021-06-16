package json

import "github.com/fullstaq-labs/sqedule/server/dbmodels"

//
// ******** Types, constants & variables ********
//

type Organization struct {
	ID          *string `json:"id"`
	DisplayName *string `json:"display_name"`
}

//
// ******** Constructor functions ********
//

func CreateFromDbOrganization(organization dbmodels.Organization) Organization {
	return Organization{
		ID:          &organization.ID,
		DisplayName: &organization.DisplayName,
	}
}

//
// ******** Other functions ********
//

func PatchDbOrganization(organization *dbmodels.Organization, json Organization) {
	if json.ID != nil {
		organization.ID = *json.ID
	}
	if json.DisplayName != nil {
		organization.DisplayName = *json.DisplayName
	}
}
