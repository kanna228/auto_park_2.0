package repository

import (
	"context"
	"database/sql"
	"fmt"
)

func (r *tripsheetRepo) Delete(ctx context.Context, id int64) error {
	return r.DeleteTripsWithTripsheet(ctx, id)
}

func (r *tripsheetRepo) DeleteTripsWithTripsheet(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx delete tripsheet: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	deleteTripsQuery := fmt.Sprintf(`
		DELETE FROM %s
		WHERE tripsheet_id = $1
	`, r.tripsheetTripsTable())

	if _, err := tx.ExecContext(ctx, deleteTripsQuery, id); err != nil {
		return fmt.Errorf("delete tripsheet trips by tripsheet id: %w", err)
	}

	deleteTripsheetQuery := fmt.Sprintf(`
		DELETE FROM %s
		WHERE id = $1
	`, r.tripsheetsTable())

	res, err := tx.ExecContext(ctx, deleteTripsheetQuery, id)
	if err != nil {
		return fmt.Errorf("delete tripsheet: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected after delete: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete tripsheet: %w", err)
	}

	return nil
}
