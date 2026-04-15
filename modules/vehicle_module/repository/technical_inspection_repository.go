package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"auto_park/modules/vehicle_module/models"
)

type TechnicalInspectionRepository interface {
	Create(ctx context.Context, p CreateTechnicalInspectionParams) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.TechnicalInspection, error)
	List(ctx context.Context, q ListTechnicalInspectionParams) ([]models.TechnicalInspection, int64, error)
	UpdateByID(ctx context.Context, id int64, p UpdateTechnicalInspectionParams) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
	UpdateFilePath(ctx context.Context, id int64, filePath string) (bool, error)
}

type CreateTechnicalInspectionParams struct {
	VehicleID int64
	Name      string
	StartDate string
	EndDate   string
	IsActive  bool
}

type UpdateTechnicalInspectionParams struct {
	VehicleID int64
	Name      string
	StartDate string
	EndDate   string
	IsActive  bool
}

type ListTechnicalInspectionParams struct {
	VehicleID *int64
	IsActive  *bool
	Name      string
	Limit     int
	Offset    int
	SortBy    string
	Order     string
}

type technicalInspectionRepoImpl struct {
	db *sql.DB
}

func NewTechnicalInspectionRepository(db *sql.DB) TechnicalInspectionRepository {
	return &technicalInspectionRepoImpl{db: db}
}

func (r *technicalInspectionRepoImpl) Create(ctx context.Context, p CreateTechnicalInspectionParams) (int64, error) {
	const q = `
		INSERT INTO technical_inspection (
			vehicle_id,
			name,
			start_date,
			end_date,
			file_path,
			is_active
		) VALUES ($1, $2, $3, $4, NULL, $5)
		RETURNING id;
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		q,
		p.VehicleID,
		p.Name,
		p.StartDate,
		p.EndDate,
		p.IsActive,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create technical inspection: %w", err)
	}

	return id, nil
}

func (r *technicalInspectionRepoImpl) GetByID(ctx context.Context, id int64) (*models.TechnicalInspection, error) {
	const q = `
		SELECT
			id,
			vehicle_id,
			name,
			start_date,
			end_date,
			file_path,
			is_active,
			created_at,
			updated_at
		FROM technical_inspection
		WHERE id = $1;
	`

	var item models.TechnicalInspection
	var filePath sql.NullString

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&item.ID,
		&item.VehicleID,
		&item.Name,
		&item.StartDate,
		&item.EndDate,
		&filePath,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get technical inspection by id: %w", err)
	}

	if filePath.Valid {
		item.FilePath = filePath.String
	}

	return &item, nil
}

func (r *technicalInspectionRepoImpl) List(ctx context.Context, q ListTechnicalInspectionParams) ([]models.TechnicalInspection, int64, error) {
	where := make([]string, 0, 4)
	args := make([]any, 0, 8)

	add := func(cond string, val any) {
		args = append(args, val)
		where = append(where, fmt.Sprintf(cond, len(args)))
	}

	if q.VehicleID != nil {
		add("vehicle_id = $%d", *q.VehicleID)
	}
	if q.IsActive != nil {
		add("is_active = $%d", *q.IsActive)
	}
	if strings.TrimSpace(q.Name) != "" {
		add("name ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.Name))
	}

	whereSQL := "TRUE"
	if len(where) > 0 {
		whereSQL = strings.Join(where, " AND ")
	}

	sortCol := "id"
	switch q.SortBy {
	case "id":
		sortCol = "id"
	case "vehicle_id":
		sortCol = "vehicle_id"
	case "name":
		sortCol = "name"
	case "start_date":
		sortCol = "start_date"
	case "end_date":
		sortCol = "end_date"
	case "created_at":
		sortCol = "created_at"
	}

	order := "DESC"
	if strings.EqualFold(q.Order, "asc") {
		order = "ASC"
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	countSQL := `SELECT COUNT(*) FROM technical_inspection WHERE ` + whereSQL
	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list technical inspection count: %w", err)
	}

	argsData := append([]any{}, args...)
	argsData = append(argsData, limit, offset)

	dataSQL := fmt.Sprintf(`
		SELECT
			id,
			vehicle_id,
			name,
			start_date,
			end_date,
			file_path,
			is_active,
			created_at,
			updated_at
		FROM technical_inspection
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortCol, order, len(args)+1, len(args)+2)

	rows, err := r.db.QueryContext(ctx, dataSQL, argsData...)
	if err != nil {
		return nil, 0, fmt.Errorf("list technical inspection query: %w", err)
	}
	defer rows.Close()

	items := make([]models.TechnicalInspection, 0, limit)
	for rows.Next() {
		var item models.TechnicalInspection
		var filePath sql.NullString

		if err := rows.Scan(
			&item.ID,
			&item.VehicleID,
			&item.Name,
			&item.StartDate,
			&item.EndDate,
			&filePath,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("list technical inspection scan: %w", err)
		}

		if filePath.Valid {
			item.FilePath = filePath.String
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list technical inspection rows: %w", err)
	}

	return items, total, nil
}

func (r *technicalInspectionRepoImpl) UpdateByID(ctx context.Context, id int64, p UpdateTechnicalInspectionParams) (bool, error) {
	const q = `
		UPDATE technical_inspection
		SET
			vehicle_id = $1,
			name = $2,
			start_date = $3,
			end_date = $4,
			is_active = $5,
			updated_at = NOW()
		WHERE id = $6;
	`

	res, err := r.db.ExecContext(
		ctx,
		q,
		p.VehicleID,
		p.Name,
		p.StartDate,
		p.EndDate,
		p.IsActive,
		id,
	)
	if err != nil {
		return false, fmt.Errorf("update technical inspection by id: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update technical inspection rows affected: %w", err)
	}

	return aff > 0, nil
}

func (r *technicalInspectionRepoImpl) DeleteByID(ctx context.Context, id int64) (bool, error) {
	const q = `DELETE FROM technical_inspection WHERE id = $1;`

	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return false, fmt.Errorf("delete technical inspection by id: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete technical inspection rows affected: %w", err)
	}

	return aff > 0, nil
}

func (r *technicalInspectionRepoImpl) UpdateFilePath(ctx context.Context, id int64, filePath string) (bool, error) {
	const q = `
		UPDATE technical_inspection
		SET
			file_path = $1,
			updated_at = NOW()
		WHERE id = $2;
	`

	var value any
	if strings.TrimSpace(filePath) == "" {
		value = nil
	} else {
		value = strings.TrimSpace(filePath)
	}

	res, err := r.db.ExecContext(ctx, q, value, id)
	if err != nil {
		return false, fmt.Errorf("update technical inspection file_path: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update technical inspection file_path rows affected: %w", err)
	}

	return aff > 0, nil
}
