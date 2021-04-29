package dbmodels

import (
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"
)

// Organization ...
type Organization struct {
	ID          string `gorm:"type:citext; primaryKey; not null"`
	DisplayName string `gorm:"not null"`
}

// FindOrganizationByID looks up an Organization by its ID.
// When not found, returns a `gorm.ErrRecordNotFound` error.
func FindOrganizationByID(db *gorm.DB, id string) (Organization, error) {
	var result Organization

	tx := db.Where("id = ?", id)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}
