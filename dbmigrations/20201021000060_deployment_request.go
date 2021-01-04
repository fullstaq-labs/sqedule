package dbmigrations

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000060)
}

var migration20201021000060 = gormigrate.Migration{
	ID: "20201021000060 Deployment request",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type Application struct {
			BaseModel
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type DeploymentRequest struct {
			BaseModel
			ApplicationID  string      `gorm:"type:citext; primaryKey; not null"`
			Application    Application `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
			ID             uint64      `gorm:"primaryKey; not null"`
			State          string      `gorm:"type:deployment_request_state; not null"`
			SourceIdentity sql.NullString
			Comments       sql.NullString
			CreatedAt      time.Time `gorm:"not null"`
			UpdatedAt      time.Time `gorm:"not null"`
			FinalizedAt    sql.NullTime
		}

		err := tx.Exec("CREATE TYPE deployment_request_state AS ENUM " +
			"('in_progress', 'cancelled', 'approved', 'rejected')").Error
		if err != nil {
			return err
		}

		err = tx.AutoMigrate(&DeploymentRequest{})
		if err != nil {
			return err
		}

		return tx.Exec("CREATE INDEX deployment_requests_created_at_idx" +
			" ON deployment_requests (organization_id, created_at DESC)").Error
	},
	Rollback: func(tx *gorm.DB) error {
		err := tx.Migrator().DropTable("deployment_requests")
		if err != nil {
			return err
		}

		return tx.Exec("DROP TYPE deployment_request_state").Error
	},
}
