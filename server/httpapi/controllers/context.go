package controllers

import (
	"sync"

	"gorm.io/gorm"
)

type Context struct {
	Db                             *gorm.DB
	AutoProcessReleaseInBackground bool
	WaitGroup                      *sync.WaitGroup
}

func NewContext(db *gorm.DB, wg *sync.WaitGroup) Context {
	return Context{
		Db:                             db,
		AutoProcessReleaseInBackground: true,
		WaitGroup:                      wg,
	}
}
