package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"auto_park/modules/notification_module/models"
)

var ErrNotificationTypeNotFound = errors.New("notification type not found")
var ErrNotificationUserNotFound = errors.New("notification user not found")

type CreateNotificationParams struct {
	UserID   int64
	TypeCode string
	Title    string
	Message  string
	Context  map[string]any
	DedupKey *string
}

type ListNotificationsParams struct {
	UserID     int64
	OnlyUnread bool
	Limit      int
	Offset     int
}

type NotificationRepository interface {
	Create(ctx context.Context, p CreateNotificationParams) (*models.Notification, error)
	ListByUser(ctx context.Context, p ListNotificationsParams) ([]models.Notification, int64, error)
	CountUnread(ctx context.Context, userID int64) (int64, error)
	MarkAsRead(ctx context.Context, userID int64, notificationID int64) (bool, error)
	MarkAllAsRead(ctx context.Context, userID int64) (int64, error)
	GetUserIDsByRole(ctx context.Context, roleCode string, fallbackRoleID int64) ([]int64, error)
}

type notificationRepo struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) Create(ctx context.Context, p CreateNotificationParams) (*models.Notification, error) {
	contextJSON, err := json.Marshal(p.Context)
	if err != nil {
		return nil, fmt.Errorf("marshal notification context: %w", err)
	}

	var dedupKey any
	if p.DedupKey != nil && strings.TrimSpace(*p.DedupKey) != "" {
		v := strings.TrimSpace(*p.DedupKey)
		dedupKey = v
	} else {
		dedupKey = nil
	}

	const q = `
		WITH nt AS (
			SELECT id
			FROM notification_types
			WHERE code = $2
		), inserted AS (
			INSERT INTO notifications (
				user_id,
				notification_type_id,
				title,
				message,
				context,
				dedup_key,
				is_readed,
				created_at,
				updated_at
			)
			SELECT $1, nt.id, $3, $4, $5::jsonb, $6, FALSE, NOW(), NOW()
			FROM nt
			ON CONFLICT (user_id, dedup_key) WHERE dedup_key IS NOT NULL DO NOTHING
			RETURNING id
		)
		SELECT id FROM inserted;
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, q, p.UserID, p.TypeCode, p.Title, p.Message, string(contextJSON), dedupKey).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Either the notification type does not exist, or this exact notification
			// has already been created for this user. Check type existence so the
			// caller still gets a useful error when type_code is wrong.
			exists, checkErr := r.notificationTypeExists(ctx, p.TypeCode)
			if checkErr != nil {
				return nil, checkErr
			}
			if !exists {
				return nil, ErrNotificationTypeNotFound
			}
			return nil, nil
		}
		return nil, mapNotificationError(err)
	}

	item, err := r.getByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *notificationRepo) ListByUser(ctx context.Context, p ListNotificationsParams) ([]models.Notification, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	conds := []string{"n.user_id = $1"}
	args := []any{p.UserID}
	argPos := 2

	if p.OnlyUnread {
		conds = append(conds, "n.is_readed = FALSE")
	}

	whereSQL := " WHERE " + strings.Join(conds, " AND ")

	countQ := `
		SELECT COUNT(*)
		FROM notifications n
	` + whereSQL

	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list notifications count: %w", err)
	}

	listQ := fmt.Sprintf(`
		SELECT
			n.id,
			n.user_id,
			n.notification_type_id,
			nt.code,
			nt.name,
			n.title,
			n.message,
			n.context,
			n.dedup_key,
			n.is_readed,
			n.read_at,
			n.created_at,
			n.updated_at
		FROM notifications n
		INNER JOIN notification_types nt ON nt.id = n.notification_type_id
		%s
		ORDER BY n.created_at DESC, n.id DESC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	items := make([]models.Notification, 0)
	for rows.Next() {
		item, err := scanNotificationRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("list notifications scan: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list notifications rows: %w", err)
	}

	return items, total, nil
}

func (r *notificationRepo) CountUnread(ctx context.Context, userID int64) (int64, error) {
	const q = `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1
		  AND is_readed = FALSE;
	`
	var count int64
	if err := r.db.QueryRowContext(ctx, q, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count unread notifications: %w", err)
	}
	return count, nil
}

func (r *notificationRepo) MarkAsRead(ctx context.Context, userID int64, notificationID int64) (bool, error) {
	const q = `
		UPDATE notifications
		SET is_readed = TRUE,
			read_at = COALESCE(read_at, NOW()),
			updated_at = NOW()
		WHERE id = $1
		  AND user_id = $2
		  AND is_readed = FALSE;
	`
	res, err := r.db.ExecContext(ctx, q, notificationID, userID)
	if err != nil {
		return false, fmt.Errorf("mark notification as read: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("mark notification as read rows affected: %w", err)
	}
	return affected > 0, nil
}

func (r *notificationRepo) MarkAllAsRead(ctx context.Context, userID int64) (int64, error) {
	const q = `
		UPDATE notifications
		SET is_readed = TRUE,
			read_at = COALESCE(read_at, NOW()),
			updated_at = NOW()
		WHERE user_id = $1
		  AND is_readed = FALSE;
	`
	res, err := r.db.ExecContext(ctx, q, userID)
	if err != nil {
		return 0, fmt.Errorf("mark all notifications as read: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("mark all notifications as read rows affected: %w", err)
	}
	return affected, nil
}

func (r *notificationRepo) GetUserIDsByRole(ctx context.Context, roleCode string, fallbackRoleID int64) ([]int64, error) {
	roleCode = strings.TrimSpace(roleCode)
	const q = `
		SELECT DISTINCT u.id
		FROM users u
		LEFT JOIN roles r ON r.id = u.role_id
		WHERE (
			($1 <> '' AND r.name = $1)
			OR ($2 > 0 AND u.role_id = $2)
		)
		ORDER BY u.id ASC;
	`
	rows, err := r.db.QueryContext(ctx, q, roleCode, fallbackRoleID)
	if err != nil {
		return nil, fmt.Errorf("get user ids by role: %w", err)
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("get user ids by role scan: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get user ids by role rows: %w", err)
	}
	return ids, nil
}

func (r *notificationRepo) getByID(ctx context.Context, id int64) (*models.Notification, error) {
	const q = `
		SELECT
			n.id,
			n.user_id,
			n.notification_type_id,
			nt.code,
			nt.name,
			n.title,
			n.message,
			n.context,
			n.dedup_key,
			n.is_readed,
			n.read_at,
			n.created_at,
			n.updated_at
		FROM notifications n
		INNER JOIN notification_types nt ON nt.id = n.notification_type_id
		WHERE n.id = $1;
	`
	item, err := scanNotification(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		return nil, fmt.Errorf("get notification by id: %w", err)
	}
	return item, nil
}

func (r *notificationRepo) notificationTypeExists(ctx context.Context, code string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM notification_types WHERE code = $1);`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, code).Scan(&exists); err != nil {
		return false, fmt.Errorf("check notification type exists: %w", err)
	}
	return exists, nil
}

type notificationScanner interface {
	Scan(dest ...any) error
}

func scanNotification(scanner notificationScanner) (*models.Notification, error) {
	var item models.Notification
	var readAt sql.NullTime
	var dedupKey sql.NullString
	var contextRaw []byte

	if err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.NotificationTypeID,
		&item.TypeCode,
		&item.TypeName,
		&item.Title,
		&item.Message,
		&contextRaw,
		&dedupKey,
		&item.IsReaded,
		&readAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if len(contextRaw) == 0 {
		contextRaw = []byte("{}")
	}
	item.Context = json.RawMessage(contextRaw)
	if dedupKey.Valid {
		item.DedupKey = &dedupKey.String
	}
	if readAt.Valid {
		item.ReadAt = &readAt.Time
	}
	return &item, nil
}

func scanNotificationRows(rows *sql.Rows) (*models.Notification, error) {
	return scanNotification(rows)
}

func mapNotificationError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotificationTypeNotFound
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "notifications_user_id_fkey"):
		return ErrNotificationUserNotFound
	case strings.Contains(msg, "notifications_notification_type_id_fkey"):
		return ErrNotificationTypeNotFound
	default:
		return err
	}
}
