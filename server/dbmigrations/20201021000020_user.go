package dbmigrations

import (
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000020)
}

var migration20201021000020 = gormigrate.Migration{
	ID: "20201021000020 User",
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

		type User struct {
			OrganizationMember
			Email        string `gorm:"type:citext; primaryKey; not null"`
			PasswordHash string `gorm:"not null"`
			FirstName    string `gorm:"not null"`
			LastName     string `gorm:"not null"`
		}

		return tx.AutoMigrate(&User{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("users")
	},
}
