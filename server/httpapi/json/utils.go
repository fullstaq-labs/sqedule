package json

import (
	"database/sql"
	"time"
)

func getSqlStringContentsOrNil(str sql.NullString) *string {
	if str.Valid {
		return &str.String
	}
	return nil
}

func getSqlTimeContentsOrNil(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func stringPointerToSqlString(str *string) sql.NullString {
	if str == nil {
		return sql.NullString{String: "", Valid: false}
	} else {
		return sql.NullString{String: *str, Valid: true}
	}
}

func int32PointerToSqlInt32(i *int32) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{Int32: 0, Valid: false}
	} else {
		return sql.NullInt32{Int32: *i, Valid: true}
	}
}
