package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"auto_park/modules/warehouse_module/models"
)

var ErrPartsCollectionNameExists = errors.New("parts collection name already exists")
var ErrServiceTypeNameExists = errors.New("service type name already exists")
var ErrVehicleServiceTypeNotFound = errors.New("service type not found")
var ErrVehicleServicePartNotFound = errors.New("parts collection item not found")
var ErrVehicleServiceVehicleNotFound = errors.New("vehicle not found")

type PartsCollectionParams struct {
	Name        string
	Description *string
}

type ListPartsCollectionParams struct {
	Name   string
	Limit  int
	Offset int
	SortBy string
	Order  string
}

type ServiceTypeParams struct {
	Name        string
	Description *string
}

type ListServiceTypesParams struct {
	Name   string
	Limit  int
	Offset int
	SortBy string
	Order  string
}

type CreateVehicleServiceParams struct {
	TypeID      int64
	PartID      int64
	VehicleID   int64
	ServiceDate string
}

type UpdateVehicleServiceParams struct {
	TypeID      int64
	PartID      int64
	VehicleID   int64
	ServiceDate string
}

type ListVehicleServicesParams struct {
	TypeID    int64
	PartID    int64
	VehicleID int64
	TypeName  string
	PartName  string
	DateFrom  string
	DateTo    string
	Limit     int
	Offset    int
	SortBy    string
	Order     string
}

type VehicleServiceRepository interface {
	CreatePartsCollection(ctx context.Context, p PartsCollectionParams) (int64, error)
	GetPartsCollectionByID(ctx context.Context, id int64) (*models.PartsCollection, error)
	ListPartsCollection(ctx context.Context, p ListPartsCollectionParams) ([]models.PartsCollection, int64, error)
	UpdatePartsCollectionByID(ctx context.Context, id int64, p PartsCollectionParams) (bool, error)
	DeletePartsCollectionByID(ctx context.Context, id int64) (bool, error)
	PartsCollectionExists(ctx context.Context, id int64) (bool, error)

	CreateServiceType(ctx context.Context, p ServiceTypeParams) (int64, error)
	GetServiceTypeByID(ctx context.Context, id int64) (*models.ServiceType, error)
	ListServiceTypes(ctx context.Context, p ListServiceTypesParams) ([]models.ServiceType, int64, error)
	UpdateServiceTypeByID(ctx context.Context, id int64, p ServiceTypeParams) (bool, error)
	DeleteServiceTypeByID(ctx context.Context, id int64) (bool, error)
	ServiceTypeExists(ctx context.Context, id int64) (bool, error)

	VehicleExists(ctx context.Context, id int64) (bool, error)
	CreateVehicleService(ctx context.Context, p CreateVehicleServiceParams) (int64, error)
	GetVehicleServiceByID(ctx context.Context, id int64) (*models.VehicleService, error)
	ListVehicleServices(ctx context.Context, p ListVehicleServicesParams) ([]models.VehicleService, int64, error)
	UpdateVehicleServiceByID(ctx context.Context, id int64, p UpdateVehicleServiceParams) (bool, error)
	DeleteVehicleServiceByID(ctx context.Context, id int64) (bool, error)
}

type vehicleServiceRepo struct {
	db *sql.DB
}

func NewVehicleServiceRepository(db *sql.DB) VehicleServiceRepository {
	return &vehicleServiceRepo{db: db}
}

func (r *vehicleServiceRepo) CreatePartsCollection(ctx context.Context, p PartsCollectionParams) (int64, error) {
	const q = `
		INSERT INTO parts_collection (name, description, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id;
	`
	var id int64
	if err := r.db.QueryRowContext(ctx, q, p.Name, p.Description).Scan(&id); err != nil {
		return 0, mapVehicleServiceError(err)
	}
	return id, nil
}

func (r *vehicleServiceRepo) GetPartsCollectionByID(ctx context.Context, id int64) (*models.PartsCollection, error) {
	const q = `SELECT id, name, description, created_at, updated_at FROM parts_collection WHERE id = $1;`
	item, err := scanPartsCollection(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get parts collection by id: %w", err)
	}
	return item, nil
}

func (r *vehicleServiceRepo) ListPartsCollection(ctx context.Context, p ListPartsCollectionParams) ([]models.PartsCollection, int64, error) {
	limit, offset := normalizeLimitOffset(p.Limit, p.Offset)
	conds := make([]string, 0, 1)
	args := make([]any, 0, 4)
	argPos := 1
	if v := strings.TrimSpace(p.Name); v != "" {
		conds = append(conds, fmt.Sprintf("name ILIKE $%d", argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}
	whereSQL := ""
	if len(conds) > 0 {
		whereSQL = " WHERE " + strings.Join(conds, " AND ")
	}

	countQ := `SELECT COUNT(*) FROM parts_collection` + whereSQL
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list parts collection count: %w", err)
	}

	sortBy := normalizePartsCollectionSortBy(p.SortBy)
	order := normalizeOrder(p.Order)
	listQ := fmt.Sprintf(`
		SELECT id, name, description, created_at, updated_at
		FROM parts_collection
		%s
		ORDER BY %s %s, id ASC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortBy, order, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list parts collection: %w", err)
	}
	defer rows.Close()

	items := make([]models.PartsCollection, 0)
	for rows.Next() {
		item, err := scanPartsCollectionRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan parts collection: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows parts collection: %w", err)
	}
	return items, total, nil
}

func (r *vehicleServiceRepo) UpdatePartsCollectionByID(ctx context.Context, id int64, p PartsCollectionParams) (bool, error) {
	const q = `UPDATE parts_collection SET name = $1, description = $2, updated_at = NOW() WHERE id = $3;`
	res, err := r.db.ExecContext(ctx, q, p.Name, p.Description, id)
	if err != nil {
		return false, mapVehicleServiceError(err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update parts collection rows affected: %w", err)
	}
	return aff > 0, nil
}

func (r *vehicleServiceRepo) DeletePartsCollectionByID(ctx context.Context, id int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM parts_collection WHERE id = $1;`, id)
	if err != nil {
		return false, mapVehicleServiceError(err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete parts collection rows affected: %w", err)
	}
	return aff > 0, nil
}

func (r *vehicleServiceRepo) PartsCollectionExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM parts_collection WHERE id = $1);`, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("check parts collection exists: %w", err)
	}
	return exists, nil
}

func (r *vehicleServiceRepo) CreateServiceType(ctx context.Context, p ServiceTypeParams) (int64, error) {
	const q = `
		INSERT INTO service_types (name, description, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id;
	`
	var id int64
	if err := r.db.QueryRowContext(ctx, q, p.Name, p.Description).Scan(&id); err != nil {
		return 0, mapVehicleServiceError(err)
	}
	return id, nil
}

func (r *vehicleServiceRepo) GetServiceTypeByID(ctx context.Context, id int64) (*models.ServiceType, error) {
	const q = `SELECT id, name, description, created_at, updated_at FROM service_types WHERE id = $1;`
	item, err := scanServiceType(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get service type by id: %w", err)
	}
	return item, nil
}

func (r *vehicleServiceRepo) ListServiceTypes(ctx context.Context, p ListServiceTypesParams) ([]models.ServiceType, int64, error) {
	limit, offset := normalizeLimitOffset(p.Limit, p.Offset)
	conds := make([]string, 0, 1)
	args := make([]any, 0, 4)
	argPos := 1
	if v := strings.TrimSpace(p.Name); v != "" {
		conds = append(conds, fmt.Sprintf("name ILIKE $%d", argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}
	whereSQL := ""
	if len(conds) > 0 {
		whereSQL = " WHERE " + strings.Join(conds, " AND ")
	}

	countQ := `SELECT COUNT(*) FROM service_types` + whereSQL
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list service types count: %w", err)
	}

	sortBy := normalizeServiceTypeSortBy(p.SortBy)
	order := normalizeOrder(p.Order)
	listQ := fmt.Sprintf(`
		SELECT id, name, description, created_at, updated_at
		FROM service_types
		%s
		ORDER BY %s %s, id ASC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortBy, order, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list service types: %w", err)
	}
	defer rows.Close()

	items := make([]models.ServiceType, 0)
	for rows.Next() {
		item, err := scanServiceTypeRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan service type: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows service types: %w", err)
	}
	return items, total, nil
}

func (r *vehicleServiceRepo) UpdateServiceTypeByID(ctx context.Context, id int64, p ServiceTypeParams) (bool, error) {
	const q = `UPDATE service_types SET name = $1, description = $2, updated_at = NOW() WHERE id = $3;`
	res, err := r.db.ExecContext(ctx, q, p.Name, p.Description, id)
	if err != nil {
		return false, mapVehicleServiceError(err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update service type rows affected: %w", err)
	}
	return aff > 0, nil
}

func (r *vehicleServiceRepo) DeleteServiceTypeByID(ctx context.Context, id int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM service_types WHERE id = $1;`, id)
	if err != nil {
		return false, mapVehicleServiceError(err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete service type rows affected: %w", err)
	}
	return aff > 0, nil
}

func (r *vehicleServiceRepo) ServiceTypeExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM service_types WHERE id = $1);`, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("check service type exists: %w", err)
	}
	return exists, nil
}

func (r *vehicleServiceRepo) VehicleExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1);`, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("check vehicle exists: %w", err)
	}
	return exists, nil
}

func (r *vehicleServiceRepo) CreateVehicleService(ctx context.Context, p CreateVehicleServiceParams) (int64, error) {
	const q = `
		INSERT INTO vehicle_services (type_id, part_id, vehicle_id, service_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id;
	`
	var id int64
	if err := r.db.QueryRowContext(ctx, q, p.TypeID, p.PartID, p.VehicleID, p.ServiceDate).Scan(&id); err != nil {
		return 0, mapVehicleServiceError(err)
	}
	return id, nil
}

func (r *vehicleServiceRepo) GetVehicleServiceByID(ctx context.Context, id int64) (*models.VehicleService, error) {
	const q = vehicleServiceSelectSQL + ` WHERE vs.id = $1;`
	item, err := scanVehicleService(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get vehicle service by id: %w", err)
	}
	return item, nil
}

func (r *vehicleServiceRepo) ListVehicleServices(ctx context.Context, p ListVehicleServicesParams) ([]models.VehicleService, int64, error) {
	limit, offset := normalizeLimitOffset(p.Limit, p.Offset)
	conds := make([]string, 0, 8)
	args := make([]any, 0, 12)
	argPos := 1
	if p.TypeID > 0 {
		conds = append(conds, fmt.Sprintf("vs.type_id = $%d", argPos))
		args = append(args, p.TypeID)
		argPos++
	}
	if p.PartID > 0 {
		conds = append(conds, fmt.Sprintf("vs.part_id = $%d", argPos))
		args = append(args, p.PartID)
		argPos++
	}
	if p.VehicleID > 0 {
		conds = append(conds, fmt.Sprintf("vs.vehicle_id = $%d", argPos))
		args = append(args, p.VehicleID)
		argPos++
	}
	if v := strings.TrimSpace(p.TypeName); v != "" {
		conds = append(conds, fmt.Sprintf("st.name ILIKE $%d", argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}
	if v := strings.TrimSpace(p.PartName); v != "" {
		conds = append(conds, fmt.Sprintf("pc.name ILIKE $%d", argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}
	if v := strings.TrimSpace(p.DateFrom); v != "" {
		conds = append(conds, fmt.Sprintf("vs.service_date >= $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := strings.TrimSpace(p.DateTo); v != "" {
		conds = append(conds, fmt.Sprintf("vs.service_date <= $%d", argPos))
		args = append(args, v)
		argPos++
	}
	whereSQL := ""
	if len(conds) > 0 {
		whereSQL = " WHERE " + strings.Join(conds, " AND ")
	}

	countQ := `
		SELECT COUNT(*)
		FROM vehicle_services vs
		INNER JOIN service_types st ON st.id = vs.type_id
		INNER JOIN parts_collection pc ON pc.id = vs.part_id
		INNER JOIN vehicles v ON v.id = vs.vehicle_id
	` + whereSQL
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list vehicle services count: %w", err)
	}

	sortBy := normalizeVehicleServiceSortBy(p.SortBy)
	order := normalizeOrder(p.Order)
	listQ := fmt.Sprintf(`
		%s
		%s
		ORDER BY %s %s, vs.id DESC
		LIMIT $%d OFFSET $%d;
	`, vehicleServiceSelectSQL, whereSQL, sortBy, order, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list vehicle services: %w", err)
	}
	defer rows.Close()

	items := make([]models.VehicleService, 0)
	for rows.Next() {
		item, err := scanVehicleServiceRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan vehicle service: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows vehicle services: %w", err)
	}
	return items, total, nil
}

func (r *vehicleServiceRepo) UpdateVehicleServiceByID(ctx context.Context, id int64, p UpdateVehicleServiceParams) (bool, error) {
	const q = `
		UPDATE vehicle_services
		SET type_id = $1,
			part_id = $2,
			vehicle_id = $3,
			service_date = $4,
			updated_at = NOW()
		WHERE id = $5;
	`
	res, err := r.db.ExecContext(ctx, q, p.TypeID, p.PartID, p.VehicleID, p.ServiceDate, id)
	if err != nil {
		return false, mapVehicleServiceError(err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update vehicle service rows affected: %w", err)
	}
	return aff > 0, nil
}

func (r *vehicleServiceRepo) DeleteVehicleServiceByID(ctx context.Context, id int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM vehicle_services WHERE id = $1;`, id)
	if err != nil {
		return false, mapVehicleServiceError(err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete vehicle service rows affected: %w", err)
	}
	return aff > 0, nil
}

const vehicleServiceSelectSQL = `
	SELECT
		vs.id,
		vs.type_id,
		st.name AS type_name,
		st.description AS type_description,
		vs.part_id,
		pc.name AS part_name,
		pc.description AS part_description,
		vs.vehicle_id,
		v.state_number AS vehicle_state_number,
		v.brand_model AS vehicle_brand_model,
		vs.service_date,
		vs.created_at,
		vs.updated_at
	FROM vehicle_services vs
	INNER JOIN service_types st ON st.id = vs.type_id
	INNER JOIN parts_collection pc ON pc.id = vs.part_id
	INNER JOIN vehicles v ON v.id = vs.vehicle_id
`

type simpleScanner interface {
	Scan(dest ...any) error
}

func scanPartsCollection(scanner simpleScanner) (*models.PartsCollection, error) {
	var item models.PartsCollection
	var description sql.NullString
	if err := scanner.Scan(&item.ID, &item.Name, &description, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	item.Description = nullableStringPtrWarehouse(description)
	return &item, nil
}

func scanPartsCollectionRows(rows *sql.Rows) (*models.PartsCollection, error) {
	return scanPartsCollection(rows)
}

func scanServiceType(scanner simpleScanner) (*models.ServiceType, error) {
	var item models.ServiceType
	var description sql.NullString
	if err := scanner.Scan(&item.ID, &item.Name, &description, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	item.Description = nullableStringPtrWarehouse(description)
	return &item, nil
}

func scanServiceTypeRows(rows *sql.Rows) (*models.ServiceType, error) {
	return scanServiceType(rows)
}

func scanVehicleService(scanner simpleScanner) (*models.VehicleService, error) {
	var item models.VehicleService
	var typeDescription sql.NullString
	var partDescription sql.NullString
	if err := scanner.Scan(
		&item.ID,
		&item.TypeID,
		&item.TypeName,
		&typeDescription,
		&item.PartID,
		&item.PartName,
		&partDescription,
		&item.VehicleID,
		&item.VehicleStateNumber,
		&item.VehicleBrandModel,
		&item.ServiceDate,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	item.TypeDescription = nullableStringPtrWarehouse(typeDescription)
	item.PartDescription = nullableStringPtrWarehouse(partDescription)
	return &item, nil
}

func scanVehicleServiceRows(rows *sql.Rows) (*models.VehicleService, error) {
	return scanVehicleService(rows)
}

func nullableStringPtrWarehouse(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	value := v.String
	return &value
}

func normalizeLimitOffset(limit int, offset int) (int, int) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func normalizePartsCollectionSortBy(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "name":
		return "name"
	case "updated_at":
		return "updated_at"
	default:
		return "created_at"
	}
}

func normalizeServiceTypeSortBy(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "name":
		return "name"
	case "updated_at":
		return "updated_at"
	default:
		return "created_at"
	}
}

func normalizeVehicleServiceSortBy(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "type_id":
		return "vs.type_id"
	case "part_id":
		return "vs.part_id"
	case "vehicle_id":
		return "vs.vehicle_id"
	case "date", "service_date":
		return "vs.service_date"
	case "updated_at":
		return "vs.updated_at"
	case "created_at":
		return "vs.created_at"
	default:
		return "vs.service_date"
	}
}

func mapVehicleServiceError(err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "parts_collection_name_key"):
		return ErrPartsCollectionNameExists
	case strings.Contains(msg, "service_types_name_key"):
		return ErrServiceTypeNameExists
	case strings.Contains(msg, "vehicle_services_type_id_fkey"):
		return ErrVehicleServiceTypeNotFound
	case strings.Contains(msg, "vehicle_services_part_id_fkey"):
		return ErrVehicleServicePartNotFound
	case strings.Contains(msg, "vehicle_services_vehicle_id_fkey"):
		return ErrVehicleServiceVehicleNotFound
	default:
		return err
	}
}
