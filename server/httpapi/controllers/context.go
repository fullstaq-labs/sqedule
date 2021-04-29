package controllers

import (
	"gorm.io/gorm"
)

type Context struct {
	Db *gorm.DB
}
