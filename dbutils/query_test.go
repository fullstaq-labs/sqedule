package dbutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUniqueConstraintError(t *testing.T) {
	db, err := SetupTestDatabase()
	if !assert.NoError(t, err) {
		return
	}

	err = db.Exec("CREATE TABLE foo (id INT PRIMARY KEY NOT NULL UNIQUE, name TEXT NOT NULL UNIQUE)").Error
	if !assert.NoError(t, err) {
		return
	}
	defer db.Exec("DROP TABLE foo")

	err = db.Exec("INSERT INTO foo VALUES(1, 'a')").Error
	if !assert.NoError(t, err) {
		return
	}

	err = db.Exec("INSERT INTO foo VALUES(1, 'b')").Error
	assert.True(t, IsUniqueConstraintError(err, "foo_pkey"), "Error=%s", err.Error())
	assert.True(t, !IsUniqueConstraintError(err, "foo_name_key"), "Error=%s", err.Error())

	err = db.Exec("INSERT INTO foo VALUES(2, 'a')").Error
	assert.True(t, !IsUniqueConstraintError(err, "foo_pkey"), "Error=%s", err.Error())
	assert.True(t, IsUniqueConstraintError(err, "foo_name_key"), "Error=%s", err.Error())
}
