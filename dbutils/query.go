package dbutils

import (
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
