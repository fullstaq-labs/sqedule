package httpapi

import (
	"gorm.io/gorm"
)

// Context ...
type Context struct {
	Db                    *gorm.DB
	UseTestAuthentication bool
}
