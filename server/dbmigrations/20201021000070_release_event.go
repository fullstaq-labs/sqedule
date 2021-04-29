package dbmigrations

import (
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000070)
}

var migration20201021000070 = gormigrate.Migration{
	ID: "20201021000070 Release event",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type Release struct {
			BaseModel
			ApplicationID string `gorm:"type:citext; primaryKey; not null"`
			ID            uint64 `gorm:"primaryKey; not null"`
		}

		type ReleaseEvent struct {
			BaseModel
			ID            uint64    `gorm:"primaryKey; not null"`
			ReleaseID     uint64    `gorm:"not null"`
			ApplicationID string    `gorm:"type:citext; not null"`
			Release       Release   `gorm:"foreignKey:OrganizationID,ApplicationID,ReleaseID; references:OrganizationID,ApplicationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
			CreatedAt     time.Time `gorm:"not null"`
		}

		type ReleaseCreatedEvent struct {
			ReleaseEvent
		}

		type ReleaseCancelledEvent struct {
			ReleaseEvent
		}

		type ReleaseRuleProcessedEvent struct {
			ReleaseEvent
			ResultState  string `gorm:"type:release_state; not null"`
			IgnoredError bool   `gorm:"not null"`
		}

		return tx.AutoMigrate(&ReleaseCreatedEvent{},
			&ReleaseCancelledEvent{}, &ReleaseRuleProcessedEvent{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("release_rule_processed_events",
			"release_cancelled_events", "release_created_events")
	},
}
