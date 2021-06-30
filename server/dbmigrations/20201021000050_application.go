package dbmigrations

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000050)
}

var migration20201021000050 = gormigrate.Migration{
	ID: "20201021000050 Application",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type ReviewableBase struct {
			CreatedAt time.Time `gorm:"not null"`
			UpdatedAt time.Time `gorm:"not null"`
		}

		type ReviewableVersionBase struct {
			ID            uint64       `gorm:"primaryKey; autoIncrement; not null"`
			VersionNumber *uint32      `gorm:"type:int; check:(version_number > 0)"`
			CreatedAt     time.Time    `gorm:"not null"`
			ApprovedAt    sql.NullTime `gorm:"check:((approved_at IS NULL) = (version_number IS NULL))"`
		}

		type ReviewableAdjustmentBase struct {
			AdjustmentNumber uint32 `gorm:"type:int; primaryKey; not null; check:(adjustment_number > 0)"`
			ReviewState      string `gorm:"type:review_state; not null"`
			ReviewComments   sql.NullString
			CreatedAt        time.Time `gorm:"not null"`
		}

		type Application struct {
			BaseModel
			ID string `gorm:"type:citext; primaryKey; not null"`
			ReviewableBase
		}

		type ApplicationVersion struct {
			BaseModel
			ReviewableVersionBase
			ApplicationID string      `gorm:"type:citext; not null"`
			Application   Application `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		type ApplicationAdjustment struct {
			BaseModel
			ApplicationVersionID uint64 `gorm:"primaryKey; not null"`
			ReviewableAdjustmentBase
			Enabled *bool `gorm:"not null; default:true"`

			DisplayName string `gorm:"not null"`

			ApplicationVersion ApplicationVersion `gorm:"foreignKey:OrganizationID,ApplicationVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		err := tx.AutoMigrate(&Application{}, &ApplicationVersion{},
			&ApplicationAdjustment{})
		if err != nil {
			return err
		}

		err = tx.Exec("CREATE UNIQUE INDEX application_version_idx" +
			" ON application_versions (organization_id, application_id, version_number DESC)" +
			" WHERE (version_number IS NOT NULL)").Error
		if err != nil {
			return err
		}

		// Work around bug in Gorm: Adjustment.VersionNumber shouldn't be autoincrement.
		err = tx.Exec("ALTER TABLE application_adjustments ALTER COLUMN adjustment_number DROP DEFAULT").Error
		if err != nil {
			return err
		}
		err = tx.Exec("DROP SEQUENCE application_adjustments_adjustment_number_seq").Error
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("application_adjustments",
			"application_versions", "applications")
	},
}
