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
			ID string `gorm:"type: citext; primaryKey; not null"`
		}

		type ApplicationMajorVersion struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null; index:version,unique"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			ID             uint64       `gorm:"primaryKey; autoIncrement; not null"`
			ApplicationID  string       `gorm:"type: citext; not null; index:version,unique"`
			Application    Application  `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		type ApplicationMinorVersion struct {
			BaseModel
			ApplicationMajorVersionID uint64                  `gorm:"primaryKey; not null"`
			VersionNumber             uint32                  `gorm:"primaryKey; not null"`
			ApplicationMajorVersion   ApplicationMajorVersion `gorm:"foreignKey:OrganizationID,ApplicationMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		type DeploymentRequest struct {
			BaseModel
			ID             uint64 `gorm:"primaryKey; not null"`
			State          string `gorm:"type:deployment_request_state; not null"`
			SourceIdentity sql.NullString
			Comments       sql.NullString
			CreatedAt      time.Time `gorm:"not null"`
			UpdatedAt      time.Time `gorm:"not null"`
			FinalizedAt    sql.NullTime

			ApplicationMajorVersionID     uint64                  `gorm:"not null"`
			ApplicationMinorVersionNumber uint32                  `gorm:"not null"`
			ApplicationMinorVersion       ApplicationMinorVersion `gorm:"foreignKey:OrganizationID,ApplicationMajorVersionID,ApplicationMinorVersionNumber; references:OrganizationID,ApplicationMajorVersionID,VersionNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		err := tx.Exec("CREATE TYPE deployment_request_state AS ENUM " +
			"('in_progress', 'cancelled', 'approved', 'rejected')").Error
		if err != nil {
			return err
		}

		return tx.AutoMigrate(&DeploymentRequest{})
	},
	Rollback: func(tx *gorm.DB) error {
		err := tx.Migrator().DropTable("deployment_requests")
		if err != nil {
			return err
		}

		return tx.Exec("DROP TYPE deployment_request_state").Error
	},
}
