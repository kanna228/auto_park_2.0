package auditlog

import (
	"context"
	"fmt"
	"log"
	"strings"

	auditlogservice "auto_park/modules/audit_log_module/service"
)

type Writer interface {
	Write(ctx context.Context, level, function, fromStatus, toStatus, actor, message string) error
}

func Actor(email string, fallbackID int64) string {
	email = strings.TrimSpace(email)
	if email != "" {
		return email
	}
	if fallbackID > 0 {
		return fmt.Sprintf("user:%d", fallbackID)
	}
	return "system"
}

func Write(ctx context.Context, svc *auditlogservice.Service, level, function, fromStatus, toStatus, actor, message string) {
	if svc == nil {
		return
	}
	if err := svc.Write(ctx, level, function, fromStatus, toStatus, actor, message); err != nil {
		log.Printf("[audit] write failed function=%s actor=%s: %v", function, actor, err)
	}
}

func Message(parts ...any) string {
	values := make([]string, 0, len(parts))
	for i := 0; i+1 < len(parts); i += 2 {
		key := strings.TrimSpace(fmt.Sprint(parts[i]))
		value := messageValue(parts[i+1])
		if key == "" || value == "" || value == "<nil>" {
			continue
		}
		values = append(values, key+"="+value)
	}
	return strings.Join(values, " ")
}

func messageValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case *string:
		if v == nil {
			return ""
		}
		return strings.TrimSpace(*v)
	case *int:
		if v == nil {
			return ""
		}
		return fmt.Sprint(*v)
	case *int64:
		if v == nil {
			return ""
		}
		return fmt.Sprint(*v)
	case *float64:
		if v == nil {
			return ""
		}
		return fmt.Sprint(*v)
	case *bool:
		if v == nil {
			return ""
		}
		return fmt.Sprint(*v)
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}
