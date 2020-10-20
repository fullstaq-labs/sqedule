package dbmodels

import "github.com/fullstaq-labs/sqedule/dbmodels/deploymentrequeststate"

type DeploymentRequestRuleProcessedEvent struct {
	DeploymentRequestEvent
	ResultState deploymentrequeststate.State `gorm:"type:deployment_request_state; not null"`
}
