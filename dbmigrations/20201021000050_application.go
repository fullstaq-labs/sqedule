package dbmigrations

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

var migration20201021000050 = gormigrate.Migration{
	ID: "20201021000050 Application",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type Application struct {
			BaseModel
			ID        string    `gorm:"primaryKey; not null"`
			CreatedAt time.Time `gorm:"not null"`
		}

		type ApplicationMajorVersion struct {
			OrganizationID string       `gorm:"primaryKey; not null; index:version,unique"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			ID             uint64       `gorm:"primaryKey; autoIncrement; not null"`
			ApplicationID  string       `gorm:"not null; index:version,unique"`
			VersionNumber  *uint32      `gorm:"index:version,unique"`
			CreatedAt      time.Time    `gorm:"not null"`
			UpdatedAt      time.Time    `gorm:"not null"`

			Application Application `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		type ApplicationMinorVersion struct {
			BaseModel
			ApplicationMajorVersionID uint64 `gorm:"primaryKey; not null"`
			VersionNumber             uint32 `gorm:"primaryKey; not null"`
			ReviewState               string `gorm:"type:review_state; not null"`
			ReviewComments            sql.NullString
			CreatedAt                 time.Time `gorm:"not null"`
			Enabled                   bool      `gorm:"not null; default:true"`

			DisplayName string `gorm:"not null"`

			ApplicationMajorVersion ApplicationMajorVersion `gorm:"foreignKey:OrganizationID,ApplicationMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		return tx.AutoMigrate(&Application{}, &ApplicationMajorVersion{},
			&ApplicationMinorVersion{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("application_minor_versions",
			"application_major_versions", "applications")
	},
}

func init() {
	registerDbMigration(&migration20201021000050)
}
