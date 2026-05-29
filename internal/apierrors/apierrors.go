package apierrors

import (
	"errors"
	"strings"

	"github.com/lib/pq"
)

const (
	CodeStockInsufficient        = "STOCK_INSUFFICIENT"
	CodePartRequestNotApproved   = "PART_REQUEST_NOT_APPROVED"
	CodePurchaseAlreadyConfirmed = "PURCHASE_ALREADY_CONFIRMED"
	CodeEntityHasReferences      = "ENTITY_HAS_REFERENCES"
	CodeEntityArchived           = "ENTITY_ARCHIVED"
	CodeUnknownRole              = "UNKNOWN_ROLE"
)

var ErrEntityArchived = errors.New("entity is archived")

func IsForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && string(pqErr.Code) == "23503" {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "violates foreign key constraint")
}
