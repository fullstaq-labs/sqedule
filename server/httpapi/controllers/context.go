package controllers

import (
	"gorm.io/gorm"
)

type Context struct {
	Db                             *gorm.DB
	AutoProcessReleaseInBackground bool
}

func NewContext(db *gorm.DB) Context {
	return Context{
		Db:                             db,
		AutoProcessReleaseInBackground: true,
	}
}
