package dbutils

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	// DefaultPageSize ...
	DefaultPageSize = 100
	// MaxPageSize ...
	MaxPageSize = 1000
)

// ApplyDbQueryPagination applies pagination settings on the given Gorm database
// query object. Pagination settings are taken from the Gin context (HTTP query strings).
// Returns the modified Gorm query object, or an error if settings parsing failed.
func ApplyDbQueryPagination(ginctx *gin.Context, tx *gorm.DB) (*gorm.DB, error) {
	var pageStr, perPageStr string
	var page, perPage uint32

	pageStr = ginctx.Query("page")
	if len(pageStr) == 0 {
		page = 1
	} else {
		i, err := strconv.ParseUint(pageStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Error parsing 'page' parameter: %w", err)
		}

		page = uint32(i)
		if page < 1 {
			return nil, errors.New("Error in 'page' parameter: must be at least 1")
		}
	}

	perPageStr = ginctx.Query("per_page")
	if len(perPageStr) == 0 {
		perPage = DefaultPageSize
	} else {
		i, err := strconv.ParseUint(perPageStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Error parsing 'per_page' parameter: %w", err)
		}

		perPage = uint32(i)
		if perPage > MaxPageSize {
			return nil, fmt.Errorf("Error in 'per_page' parameter: may not be larger than %d", MaxPageSize)
		}
	}

	return tx.Offset(int((page - 1) * perPage)).Limit(int(perPage)), nil
}
