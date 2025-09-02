package helpers

import (
	"jwt_go/graph/model"
	"math"
	"strings"

	"gorm.io/gorm"
)

// ApplyPagination applies sorting, limit, and offset to the given GORM query based on PaginationInput.
// It returns the updated query, along with the resolved limit and page used.
func ApplyPagination(db *gorm.DB, pagination *model.PaginationInput, defaultLimit int32, validSortFields map[string]bool) (*gorm.DB, int32, int32) {
	// Defaults
	limit := int32(10)
	if defaultLimit > 0 {
		limit = defaultLimit
	}
	page := int32(1)
	sortBy := "created_at,desc"

	if pagination != nil {
		if pagination.Limit != nil && *pagination.Limit > 0 {
			limit = *pagination.Limit
		}
		if pagination.Page != nil && *pagination.Page > 0 {
			page = *pagination.Page
		}
		if pagination.SortBy != nil && *pagination.SortBy != "" {
			sortBy = *pagination.SortBy
		}

		// Apply sorting with whitelist validation
		sortParts := strings.Split(sortBy, ",")
		if len(sortParts) == 2 {
			field := strings.TrimSpace(sortParts[0])
			order := strings.ToUpper(strings.TrimSpace(sortParts[1]))
			if validSortFields[field] && (order == "ASC" || order == "DESC") {
				db = db.Order(field + " " + order)
			} else {
				// Fallback
				db = db.Order("created_at DESC")
			}
		} else {
			// Fallback sorting
			db = db.Order("created_at DESC")
		}

		// Apply limit/offset
		offset := (page - 1) * limit
		db = db.Offset(int(offset)).Limit(int(limit))
	} else {
		// Defaults when no pagination provided
		db = db.Order("created_at DESC").Limit(int(limit))
	}

	return db, limit, page
}

// BuildPageInfo builds a PageInfo object from total items, limit, page, and the number of items in the current page.
func BuildPageInfo(totalItems int64, limit int32, page int32, itemsCount int) *model.PageInfo {
	var totalPages int64
	if limit > 0 {
		totalPages = int64(math.Ceil(float64(totalItems) / float64(limit)))
	}
	if totalPages < 0 {
		totalPages = 0
	}

	hasNext := page < int32(totalPages)
	hasPrev := page > 1

	var startItemPtr, endItemPtr *int32
	if itemsCount > 0 {
		start := (page-1)*limit + 1
		end := page * limit
		if end > int32(totalItems) {
			end = int32(totalItems)
		}
		startItemPtr = &start
		endItemPtr = &end
	}

	return &model.PageInfo{
		CurrentPage:     page,
		PerPage:         limit,
		TotalItems:      int32(totalItems),
		TotalPages:      int32(totalPages),
		HasNextPage:     hasNext,
		HasPreviousPage: hasPrev,
		StartItem:       startItemPtr,
		EndItem:         endItemPtr,
	}
}
