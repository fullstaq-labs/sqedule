package httpapi

import (
	"sync"

	"gorm.io/gorm"
)

type Context struct {
	Db                    *gorm.DB
	WaitGroup             *sync.WaitGroup
	UseTestAuthentication bool
	DevelopmentMode       bool
	CorsOrigin            string
}
