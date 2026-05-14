package middleware

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPageSize = 50
	MaxPageSize     = 200
)

// PaginationAliasMiddleware lets all existing list handlers support both styles:
//
//	?limit=50&offset=0
//	?page=1&page_size=50
//
// If limit/offset are already provided, they stay untouched.
func PaginationAliasMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Request.URL.Query()

		if strings.TrimSpace(q.Get("limit")) == "" {
			pageSize := parsePositiveInt(q.Get("page_size"), DefaultPageSize)
			if pageSize > MaxPageSize {
				pageSize = MaxPageSize
			}
			q.Set("limit", strconv.Itoa(pageSize))
		}

		if strings.TrimSpace(q.Get("offset")) == "" {
			page := parsePositiveInt(q.Get("page"), 1)
			limit := parsePositiveInt(q.Get("limit"), DefaultPageSize)
			if limit > MaxPageSize {
				limit = MaxPageSize
			}
			offset := (page - 1) * limit
			q.Set("offset", strconv.Itoa(offset))
		}

		c.Request.URL.RawQuery = encodeQuery(q)
		c.Next()
	}
}

func parsePositiveInt(value string, fallback int) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func encodeQuery(q url.Values) string {
	return q.Encode()
}
