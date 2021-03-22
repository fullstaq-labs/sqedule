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

type PaginationOptions struct {
	Page    uint
	PerPage uint
}

func ParsePaginationOptions(ginctx *gin.Context) (PaginationOptions, error) {
	var options PaginationOptions
	var pageStr, perPageStr string

	pageStr = ginctx.Query("page")
	if len(pageStr) == 0 {
		options.Page = 1
	} else {
		i, err := strconv.ParseUint(pageStr, 10, 32)
		if err != nil {
			return PaginationOptions{}, fmt.Errorf("Error parsing 'page' parameter: %w", err)
		}

		options.Page = uint(i)
		if options.Page < 1 {
			return PaginationOptions{}, errors.New("Error in 'page' parameter: must be at least 1")
		}
	}

	perPageStr = ginctx.Query("per_page")
	if len(perPageStr) == 0 {
		options.PerPage = DefaultPageSize
	} else {
		i, err := strconv.ParseUint(perPageStr, 10, 32)
		if err != nil {
			return PaginationOptions{}, fmt.Errorf("Error parsing 'per_page' parameter: %w", err)
		}

		options.PerPage = uint(i)
		if options.PerPage > MaxPageSize {
			return PaginationOptions{}, fmt.Errorf("Error in 'per_page' parameter: may not be larger than %d", MaxPageSize)
		}
	}

	return options, nil
}

// ApplyDbQueryPagination applies pagination settings on the given Gorm database
// query object. Pagination settings are taken from the Gin context (HTTP query strings).
// Returns the modified Gorm query object, or an error if settings parsing failed.
func ApplyDbQueryPagination(ginctx *gin.Context, tx *gorm.DB) (*gorm.DB, error) {
	options, err := ParsePaginationOptions(ginctx)
	if err != nil {
		return nil, err
	}

	return ApplyDbQueryPaginationOptions(tx, options), nil
}

func ApplyDbQueryPaginationOptions(tx *gorm.DB, options PaginationOptions) *gorm.DB {
	return tx.Offset(int((options.Page - 1) * options.PerPage)).Limit(int(options.PerPage))
}
