package dbmigrations

import (
	"time"

	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000030)
}

var migration20201021000030 = gormigrate.Migration{
	ID: "20201021000030 Service account",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type OrganizationMember struct {
			BaseModel
			Role      string    `gorm:"type:organization_member_role; not null"`
			CreatedAt time.Time `gorm:"not null"`
			UpdatedAt time.Time `gorm:"not null"`
		}

		type ServiceAccount struct {
			OrganizationMember
			Name       string `gorm:"type:citext; primaryKey; not null"`
			SecretHash string `gorm:"not null"`
		}

		return tx.AutoMigrate(&ServiceAccount{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("service_accounts")
	},
}
