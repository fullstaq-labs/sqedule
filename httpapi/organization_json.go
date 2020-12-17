package httpapi

import "github.com/fullstaq-labs/sqedule/dbmodels"

type organizationJSON struct {
	ID          *string `json:"id"`
	DisplayName *string `json:"display_name"`
}

func createOrganizationJSONFromDbModel(organization dbmodels.Organization) organizationJSON {
	return organizationJSON{
		ID:          &organization.ID,
		DisplayName: &organization.DisplayName,
	}
}

func patchOrganizationDbModelFromJSON(organization *dbmodels.Organization, json organizationJSON) {
	if json.ID != nil {
		organization.ID = *json.ID
	}
	if json.DisplayName != nil {
		organization.DisplayName = *json.DisplayName
	}
}
