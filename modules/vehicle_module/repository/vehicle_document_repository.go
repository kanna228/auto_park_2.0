package repository

import (
	"context"
	"database/sql"
	"fmt"

	"auto_park/modules/vehicle_module/models"
)

type VehicleDocumentRepository interface {
	ListByVehicle(ctx context.Context, vehicleID int64) ([]models.VehicleDocument, error)
	Create(ctx context.Context, vehicleID int64, p CreateVehicleDocumentParams) (int64, error)
	Delete(ctx context.Context, vehicleID int64, documentID int64) (bool, error)
}

type CreateVehicleDocumentParams struct {
	Type      string
	Number    string
	ValidFrom string
	ValidTo   string
}

type vehicleDocumentRepo struct {
	db *sql.DB
}

func NewVehicleDocumentRepository(db *sql.DB) VehicleDocumentRepository {
	return &vehicleDocumentRepo{db: db}
}

func (r *vehicleDocumentRepo) ListByVehicle(ctx context.Context, vehicleID int64) ([]models.VehicleDocument, error) {
	const q = `
		SELECT id, vehicle_id, type, number, valid_from, valid_to, created_at, updated_at
		FROM vehicle_documents
		WHERE vehicle_id = $1
		ORDER BY valid_from DESC, id DESC;
	`

	rows, err := r.db.QueryContext(ctx, q, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("list vehicle documents: %w", err)
	}
	defer rows.Close()

	items := make([]models.VehicleDocument, 0)
	for rows.Next() {
		var item models.VehicleDocument
		if err := rows.Scan(
			&item.ID,
			&item.VehicleID,
			&item.Type,
			&item.Number,
			&item.ValidFrom,
			&item.ValidTo,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("list vehicle documents scan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list vehicle documents rows: %w", err)
	}
	return items, nil
}

func (r *vehicleDocumentRepo) Create(ctx context.Context, vehicleID int64, p CreateVehicleDocumentParams) (int64, error) {
	const q = `
		INSERT INTO vehicle_documents (vehicle_id, type, number, valid_from, valid_to, created_at, updated_at)
		VALUES ($1, $2, $3, $4::date, $5::date, NOW(), NOW())
		RETURNING id;
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, q, vehicleID, p.Type, p.Number, p.ValidFrom, p.ValidTo).Scan(&id); err != nil {
		return 0, fmt.Errorf("create vehicle document: %w", err)
	}
	return id, nil
}

func (r *vehicleDocumentRepo) Delete(ctx context.Context, vehicleID int64, documentID int64) (bool, error) {
	const q = `
		DELETE FROM vehicle_documents
		WHERE id = $1
		  AND vehicle_id = $2;
	`

	res, err := r.db.ExecContext(ctx, q, documentID, vehicleID)
	if err != nil {
		return false, fmt.Errorf("delete vehicle document: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete vehicle document rows affected: %w", err)
	}
	return affected > 0, nil
}
