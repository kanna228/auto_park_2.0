package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/audit_log_module/dto"
	"auto_park/modules/audit_log_module/repository"
)

type Service struct {
	repo repository.AuditLogRepository
}

func NewAuditLogService(repo repository.AuditLogRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Write(ctx context.Context, level, function, fromStatus, toStatus, actor, message string) error {
	level = strings.ToLower(strings.TrimSpace(level))
	function = strings.ToLower(strings.TrimSpace(function))
	actor = strings.TrimSpace(actor)
	if actor == "" {
		actor = "system"
	}

	if !validAuditLevel(level) {
		return fmt.Errorf("invalid level")
	}
	if !validAuditFunction(function) {
		return fmt.Errorf("invalid function")
	}

	return s.repo.Create(ctx, repository.CreateAuditLogParams{
		Level:      level,
		Function:   function,
		FromStatus: optionalString(fromStatus),
		ToStatus:   optionalString(toStatus),
		Actor:      actor,
		Message:    optionalString(message),
	})
}

func (s *Service) List(ctx context.Context, q dto.AuditLogListQuery) (*dto.AuditLogListResponse, error) {
	if err := validateAuditQuery(q); err != nil {
		return nil, err
	}
	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, err
	}
	return &dto.AuditLogListResponse{
		Items:  items,
		Total:  total,
		Limit:  normalizeLimit(q.Limit),
		Offset: normalizeOffset(q.Offset),
	}, nil
}

func (s *Service) ExportCSV(ctx context.Context, q dto.AuditLogListQuery) ([]byte, error) {
	if err := validateAuditQuery(q); err != nil {
		return nil, err
	}
	q.Limit = 10000
	q.Offset = 0
	items, _, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"id", "level", "function", "from_status", "to_status", "actor", "message", "created_at"}); err != nil {
		return nil, err
	}
	for _, item := range items {
		if err := w.Write([]string{
			fmt.Sprint(item.ID),
			item.Level,
			item.Function,
			deref(item.FromStatus),
			deref(item.ToStatus),
			item.Actor,
			deref(item.Message),
			item.CreatedAt.Format(time.RFC3339),
		}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func validateAuditQuery(q dto.AuditLogListQuery) error {
	if v := strings.TrimSpace(q.Level); v != "" && !validAuditLevel(v) {
		return fmt.Errorf("invalid level")
	}
	if v := strings.TrimSpace(q.Function); v != "" && !validAuditFunction(v) {
		return fmt.Errorf("invalid function")
	}
	if v := strings.TrimSpace(q.Date); v != "" {
		if _, err := time.Parse("2006-01-02", v); err != nil {
			return fmt.Errorf("date must be in YYYY-MM-DD format")
		}
	}
	return nil
}

func validAuditLevel(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "system", "info", "success", "warning", "error":
		return true
	default:
		return false
	}
}

func validAuditFunction(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "request", "arrival", "tripsheet", "shift", "incident", "vehicle", "user":
		return true
	default:
		return false
	}
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 10000 {
		return 10000
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
