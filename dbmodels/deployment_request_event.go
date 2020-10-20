package dbmodels

import "time"

type DeploymentRequestEvent struct {
	BaseModel
	ID                  uint64            `gorm:"primaryKey; not null"`
	DeploymentRequestID uint64            `gorm:"not null"`
	DeploymentRequest   DeploymentRequest `gorm:"foreignKey:OrganizationID,DeploymentRequestID references:OrganizationID,ID"`
	CreatedAt           time.Time         `gorm:"not null"`
}
