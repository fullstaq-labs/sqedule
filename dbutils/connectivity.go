package dbutils

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UnsupportedDatabaseTypeError struct {
	Type string
}

func (e *UnsupportedDatabaseTypeError) Error() string {
	return fmt.Sprintf("Unsupported database type %s", e.Type)
}

func EstablishDatabaseConnection(dbtype string, connString string, config *gorm.Config) (*gorm.DB, error) {
	// If you add a new supported database type, be sure to update cmd/db_help.go too.
	switch dbtype {
	case "postgresql":
		return gorm.Open(postgres.Open(connString), config)
	default:
		return nil, &UnsupportedDatabaseTypeError{dbtype}
	}
}
