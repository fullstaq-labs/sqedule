package dbutils

import (
	"strings"

	"gorm.io/gorm"
)

// QueryStringList performs a query that only selects a single column of type string.
// It returns all rows (as an array of strings), or an error.
//
//     QueryStringList(db, "SELECT name FROM users WHERE age > ?", 12)
func QueryStringList(db *gorm.DB, sql string, values ...interface{}) ([]string, error) {
	var result []string

	rows, err := db.Raw(sql, values...).Rows()
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}

		result = append(result, value)
	}

	return result, nil
}

// CreateFindOperationError is to be used by `dbmodel.FindXxxByYyyy()`
// functions to ensure that, when a record is not found, a
// `gorm.ErrRecordNotFound` error is returned.
func CreateFindOperationError(tx *gorm.DB) error {
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// IsUniqueConstraintError checks whether the given gorm error represents a unique key constraint error,
// on the given constraint name.
func IsUniqueConstraintError(err error, constraintName string) bool {
	return strings.Index(err.Error(), "violates unique constraint \""+constraintName+"\"") != -1 &&
		strings.Index(err.Error(), "SQLSTATE 23505") != -1
}
