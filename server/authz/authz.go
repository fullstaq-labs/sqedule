package authz

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

type CollectionAction string
type SingularAction string

type IAuthorizer interface {
	CollectionAuthorizations(orgMember dbmodels.IOrganizationMember) map[CollectionAction]struct{}
	SingularAuthorizations(orgMember dbmodels.IOrganizationMember, target interface{}) map[SingularAction]struct{}
}

// Authorize checks whether an OrganizationMember is allowed to
// perform the given collection action.
func AuthorizeCollectionAction(authorizer IAuthorizer, orgMember dbmodels.IOrganizationMember, action CollectionAction) bool {
	permittedActions := authorizer.CollectionAuthorizations(orgMember)
	_, contains := permittedActions[action]
	return contains
}

// Authorize checks whether an OrganizationMember is allowed to
// perform the given action on a target object.
func AuthorizeSingularAction(authorizer IAuthorizer, orgMember dbmodels.IOrganizationMember, action SingularAction, target interface{}) bool {
	permittedActions := authorizer.SingularAuthorizations(orgMember, target)
	_, contains := permittedActions[action]
	return contains
}
