package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func (r *tripsheetRepo) ResolveDriverID(ctx context.Context, driverID *int64, driverShiftID *int64) (*int64, error) {
	if driverID != nil && *driverID > 0 {
		v := *driverID
		return &v, nil
	}
	if driverShiftID == nil || *driverShiftID <= 0 {
		return nil, nil
	}

	const q = `SELECT driver_id FROM driver_shifts WHERE id = $1 AND is_deleted = FALSE;`
	var id int64
	if err := r.db.QueryRowContext(ctx, q, *driverShiftID).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("resolve tripsheet driver by shift: %w", err)
	}
	return &id, nil
}

func (r *tripsheetRepo) SetDriverStatusByCode(ctx context.Context, driverID int64, code string) error {
	code = normalizeDriverStatusCode(code)
	if driverID <= 0 || code == "" {
		return nil
	}

	const q = `
		WITH target_status AS (
			SELECT id, code
			FROM driver_statuses
			WHERE code = $2
			LIMIT 1
		)
		UPDATE drivers d
		SET status_id = target_status.id,
			status = target_status.code,
			updated_at = NOW()
		FROM target_status
		WHERE d.id = $1;
	`
	if _, err := r.db.ExecContext(ctx, q, driverID, code); err != nil {
		return fmt.Errorf("set driver status by code: %w", err)
	}
	return nil
}

func (r *tripsheetRepo) RefreshDriverAvailability(ctx context.Context, driverID int64) error {
	if driverID <= 0 {
		return nil
	}

	const q = `
		WITH activity AS (
			SELECT EXISTS (
				SELECT 1
				FROM tripsheets t
				LEFT JOIN driver_shifts ds ON ds.id = t.driver_shift_id
				WHERE COALESCE(t.driver_id, ds.driver_id) = $1
				  AND t.end_time IS NULL
				  AND t.status_id NOT IN (4, 5)
			) AS has_activity
		), target_status AS (
			SELECT id, code
			FROM driver_statuses, activity
			WHERE code = CASE WHEN activity.has_activity THEN 'on_trip' ELSE 'available' END
			LIMIT 1
		)
		UPDATE drivers d
		SET status_id = target_status.id,
			status = target_status.code,
			updated_at = NOW()
		FROM target_status, activity
		WHERE d.id = $1
		  AND (activity.has_activity OR d.status = 'on_trip');
	`
	if _, err := r.db.ExecContext(ctx, q, driverID); err != nil {
		return fmt.Errorf("refresh driver availability: %w", err)
	}
	return nil
}

func normalizeDriverStatusCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	switch code {
	case "available", "on_trip", "inactive":
		return code
	case "unavailable":
		return "inactive"
	default:
		return code
	}
}
