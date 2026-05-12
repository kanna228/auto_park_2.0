package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"auto_park/modules/warehouse_module/models"
)

var ErrVehiclePartInstallationPartNotFound = errors.New("part not found")
var ErrVehiclePartInstallationVehicleNotFound = errors.New("vehicle not found")
var ErrVehiclePartInstallationUserNotFound = errors.New("installer user not found")
var ErrVehiclePartInstallationMechanicShiftNotFound = errors.New("mechanic shift not found")
var ErrVehiclePartInstallationInsufficientStock = errors.New("not enough part quantity in warehouse")
var ErrVehiclePartInstallationActiveDuplicate = errors.New("non-consumable part is already active on this vehicle")

type CreateVehiclePartInstallationParams struct {
	PartID               int64
	VehicleID            int64
	MechanicShiftID      int64
	InstalledAt          string
	PlannedReplacementAt string
	Quantity             int64
	InstalledByUserID    int64
}

type UpdateVehiclePartInstallationParams struct {
	PartID               int64
	VehicleID            int64
	MechanicShiftID      int64
	InstalledAt          string
	PlannedReplacementAt string
	Quantity             int64
	InstalledByUserID    int64
	IsActive             bool
}

type ListVehiclePartInstallationsParams struct {
	PartID            int64
	VehicleID         int64
	MechanicShiftID   int64
	InstalledByUserID int64
	IsActive          *bool
	DateFrom          string
	DateTo            string
	ReplacementFrom   string
	ReplacementTo     string
	Limit             int
	Offset            int
	SortBy            string
	Order             string
}

type VehiclePartInstallationRepository interface {
	Create(ctx context.Context, p CreateVehiclePartInstallationParams) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.VehiclePartInstallation, error)
	List(ctx context.Context, p ListVehiclePartInstallationsParams) ([]models.VehiclePartInstallation, int64, error)
	UpdateByID(ctx context.Context, id int64, p UpdateVehiclePartInstallationParams) (bool, error)
	UpdateActivityByID(ctx context.Context, id int64, isActive bool) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
}

type vehiclePartInstallationRepo struct {
	db *sql.DB
}

func NewVehiclePartInstallationRepository(db *sql.DB) VehiclePartInstallationRepository {
	return &vehiclePartInstallationRepo{db: db}
}

type warehousePartForInstallation struct {
	ID           int64
	PartID       string
	Name         string
	Quantity     int64
	Category     string
	IsConsumable bool
}

type vehiclePartInstallationState struct {
	ID                   int64
	PartID               int64
	VehicleID            int64
	MechanicShiftID      *int64
	InstalledAt          time.Time
	PlannedReplacementAt time.Time
	Quantity             int64
	InstalledByUserID    int64
	IsActive             bool
}

func (r *vehiclePartInstallationRepo) Create(ctx context.Context, p CreateVehiclePartInstallationParams) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin create vehicle part installation: %w", err)
	}
	defer rollbackUnlessCommitted(tx)

	part, err := getPartForUpdate(ctx, tx, p.PartID)
	if err != nil {
		return 0, err
	}

	if err := ensureVehicleExistsTx(ctx, tx, p.VehicleID); err != nil {
		return 0, err
	}
	if err := ensureInstallerExistsTx(ctx, tx, p.InstalledByUserID); err != nil {
		return 0, err
	}
	if err := ensureMechanicShiftExistsTx(ctx, tx, p.MechanicShiftID); err != nil {
		return 0, err
	}
	if part.Quantity < p.Quantity {
		return 0, ErrVehiclePartInstallationInsufficientStock
	}
	if !part.IsConsumable {
		exists, err := activeInstallationExistsTx(ctx, tx, p.PartID, p.VehicleID, 0)
		if err != nil {
			return 0, err
		}
		if exists {
			return 0, ErrVehiclePartInstallationActiveDuplicate
		}
	}

	if err := decreasePartQuantityTx(ctx, tx, p.PartID, p.Quantity); err != nil {
		return 0, err
	}

	const insertQ = `
		INSERT INTO vehicle_part_installations (
			part_id,
			vehicle_id,
			mechanic_shift_id,
			installed_at,
			planned_replacement_at,
			quantity,
			installed_by_user_id,
			is_active,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE, NOW(), NOW())
		RETURNING id;
	`

	var id int64
	if err := tx.QueryRowContext(
		ctx,
		insertQ,
		p.PartID,
		p.VehicleID,
		p.MechanicShiftID,
		p.InstalledAt,
		p.PlannedReplacementAt,
		p.Quantity,
		p.InstalledByUserID,
	).Scan(&id); err != nil {
		return 0, mapVehiclePartInstallationError(err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit create vehicle part installation: %w", err)
	}

	return id, nil
}

func (r *vehiclePartInstallationRepo) GetByID(ctx context.Context, id int64) (*models.VehiclePartInstallation, error) {
	const q = vehiclePartInstallationSelectSQL + ` WHERE vpi.id = $1;`
	item, err := scanVehiclePartInstallation(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get vehicle part installation by id: %w", err)
	}
	return item, nil
}

func (r *vehiclePartInstallationRepo) List(ctx context.Context, p ListVehiclePartInstallationsParams) ([]models.VehiclePartInstallation, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	conds := make([]string, 0, 10)
	args := make([]any, 0, 12)
	argPos := 1

	if p.PartID > 0 {
		conds = append(conds, fmt.Sprintf("vpi.part_id = $%d", argPos))
		args = append(args, p.PartID)
		argPos++
	}
	if p.VehicleID > 0 {
		conds = append(conds, fmt.Sprintf("vpi.vehicle_id = $%d", argPos))
		args = append(args, p.VehicleID)
		argPos++
	}
	if p.MechanicShiftID > 0 {
		conds = append(conds, fmt.Sprintf("vpi.mechanic_shift_id = $%d", argPos))
		args = append(args, p.MechanicShiftID)
		argPos++
	}
	if p.InstalledByUserID > 0 {
		conds = append(conds, fmt.Sprintf("vpi.installed_by_user_id = $%d", argPos))
		args = append(args, p.InstalledByUserID)
		argPos++
	}
	if p.IsActive != nil {
		conds = append(conds, fmt.Sprintf("vpi.is_active = $%d", argPos))
		args = append(args, *p.IsActive)
		argPos++
	}
	if v := strings.TrimSpace(p.DateFrom); v != "" {
		conds = append(conds, fmt.Sprintf("vpi.installed_at >= $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := strings.TrimSpace(p.DateTo); v != "" {
		conds = append(conds, fmt.Sprintf("vpi.installed_at <= $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := strings.TrimSpace(p.ReplacementFrom); v != "" {
		conds = append(conds, fmt.Sprintf("vpi.planned_replacement_at >= $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := strings.TrimSpace(p.ReplacementTo); v != "" {
		conds = append(conds, fmt.Sprintf("vpi.planned_replacement_at <= $%d", argPos))
		args = append(args, v)
		argPos++
	}

	whereSQL := ""
	if len(conds) > 0 {
		whereSQL = " WHERE " + strings.Join(conds, " AND ")
	}

	countQ := `
		SELECT COUNT(*)
		FROM vehicle_part_installations vpi
		INNER JOIN parts_catalog p ON p.id = vpi.part_id
		INNER JOIN vehicles v ON v.id = vpi.vehicle_id
		LEFT JOIN users u ON u.id = vpi.installed_by_user_id
		LEFT JOIN mechanic_shifts ms ON ms.id = vpi.mechanic_shift_id
		LEFT JOIN users msu ON msu.id = ms.user_id
	` + whereSQL

	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list vehicle part installations count: %w", err)
	}

	sortBy := normalizeVehiclePartInstallationSortBy(p.SortBy)
	order := normalizeOrder(p.Order)
	listQ := fmt.Sprintf(`
		%s
		%s
		ORDER BY %s %s, vpi.id DESC
		LIMIT $%d OFFSET $%d;
	`, vehiclePartInstallationSelectSQL, whereSQL, sortBy, order, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list vehicle part installations: %w", err)
	}
	defer rows.Close()

	items := make([]models.VehiclePartInstallation, 0)
	for rows.Next() {
		item, err := scanVehiclePartInstallationRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("list vehicle part installations scan: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list vehicle part installations rows: %w", err)
	}

	return items, total, nil
}

func (r *vehiclePartInstallationRepo) UpdateByID(ctx context.Context, id int64, p UpdateVehiclePartInstallationParams) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin update vehicle part installation: %w", err)
	}
	defer rollbackUnlessCommitted(tx)

	current, err := getInstallationForUpdate(ctx, tx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	parts, err := lockPartsForUpdate(ctx, tx, current.PartID, p.PartID)
	if err != nil {
		return false, err
	}
	newPart := parts[p.PartID]
	if newPart == nil {
		return false, ErrVehiclePartInstallationPartNotFound
	}

	if err := ensureVehicleExistsTx(ctx, tx, p.VehicleID); err != nil {
		return false, err
	}
	if err := ensureInstallerExistsTx(ctx, tx, p.InstalledByUserID); err != nil {
		return false, err
	}
	if err := ensureMechanicShiftExistsTx(ctx, tx, p.MechanicShiftID); err != nil {
		return false, err
	}

	if p.IsActive && !newPart.IsConsumable {
		exists, err := activeInstallationExistsTx(ctx, tx, p.PartID, p.VehicleID, id)
		if err != nil {
			return false, err
		}
		if exists {
			return false, ErrVehiclePartInstallationActiveDuplicate
		}
	}

	if current.PartID == p.PartID {
		delta := p.Quantity - current.Quantity
		if delta > 0 {
			if newPart.Quantity < delta {
				return false, ErrVehiclePartInstallationInsufficientStock
			}
			if err := decreasePartQuantityTx(ctx, tx, p.PartID, delta); err != nil {
				return false, err
			}
		} else if delta < 0 {
			if err := increasePartQuantityTx(ctx, tx, p.PartID, -delta); err != nil {
				return false, err
			}
		}
	} else {
		if err := increasePartQuantityTx(ctx, tx, current.PartID, current.Quantity); err != nil {
			return false, err
		}
		if newPart.Quantity < p.Quantity {
			return false, ErrVehiclePartInstallationInsufficientStock
		}
		if err := decreasePartQuantityTx(ctx, tx, p.PartID, p.Quantity); err != nil {
			return false, err
		}
	}

	const q = `
		UPDATE vehicle_part_installations
		SET part_id = $1,
			vehicle_id = $2,
			mechanic_shift_id = $3,
			installed_at = $4,
			planned_replacement_at = $5,
			quantity = $6,
			installed_by_user_id = $7,
			is_active = $8,
			updated_at = NOW()
		WHERE id = $9;
	`
	res, err := tx.ExecContext(
		ctx,
		q,
		p.PartID,
		p.VehicleID,
		p.MechanicShiftID,
		p.InstalledAt,
		p.PlannedReplacementAt,
		p.Quantity,
		p.InstalledByUserID,
		p.IsActive,
		id,
	)
	if err != nil {
		return false, mapVehiclePartInstallationError(err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update vehicle part installation rows affected: %w", err)
	}
	if aff == 0 {
		return false, nil
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit update vehicle part installation: %w", err)
	}
	return true, nil
}

func (r *vehiclePartInstallationRepo) UpdateActivityByID(ctx context.Context, id int64, isActive bool) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin update vehicle part installation activity: %w", err)
	}
	defer rollbackUnlessCommitted(tx)

	current, err := getInstallationForUpdate(ctx, tx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	part, err := getPartForUpdate(ctx, tx, current.PartID)
	if err != nil {
		return false, err
	}
	if isActive && !part.IsConsumable {
		exists, err := activeInstallationExistsTx(ctx, tx, current.PartID, current.VehicleID, id)
		if err != nil {
			return false, err
		}
		if exists {
			return false, ErrVehiclePartInstallationActiveDuplicate
		}
	}

	const q = `
		UPDATE vehicle_part_installations
		SET is_active = $1,
			updated_at = NOW()
		WHERE id = $2;
	`
	res, err := tx.ExecContext(ctx, q, isActive, id)
	if err != nil {
		return false, mapVehiclePartInstallationError(err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update vehicle part installation activity rows affected: %w", err)
	}
	if aff == 0 {
		return false, nil
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit update vehicle part installation activity: %w", err)
	}
	return true, nil
}

func (r *vehiclePartInstallationRepo) DeleteByID(ctx context.Context, id int64) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin delete vehicle part installation: %w", err)
	}
	defer rollbackUnlessCommitted(tx)

	current, err := getInstallationForUpdate(ctx, tx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	if _, err := getPartForUpdate(ctx, tx, current.PartID); err != nil {
		return false, err
	}
	if err := increasePartQuantityTx(ctx, tx, current.PartID, current.Quantity); err != nil {
		return false, err
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM vehicle_part_installations WHERE id = $1;`, id)
	if err != nil {
		return false, fmt.Errorf("delete vehicle part installation by id: %w", err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete vehicle part installation rows affected: %w", err)
	}
	if aff == 0 {
		return false, nil
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit delete vehicle part installation: %w", err)
	}
	return true, nil
}

const vehiclePartInstallationSelectSQL = `
	SELECT
		vpi.id,
		vpi.part_id,
		p.part_id AS part_catalog_code,
		p.name AS part_name,
		p.category AS part_category,
		p.is_consumable AS part_is_consumable,
		vpi.vehicle_id,
		v.state_number AS vehicle_state_number,
		v.brand_model AS vehicle_brand_model,
		vpi.mechanic_shift_id,
		ms.user_id AS mechanic_shift_user_id,
		ms.shift_date AS mechanic_shift_date,
		ms.time_from AS mechanic_shift_time_from,
		ms.time_to AS mechanic_shift_time_to,
		msu.email AS mechanic_shift_user_email,
		NULLIF(TRIM(CONCAT_WS(' ', msu.last_name, msu.first_name, msu.middle_name)), '') AS mechanic_shift_user_full_name,
		vpi.installed_at,
		vpi.planned_replacement_at,
		vpi.quantity,
		vpi.installed_by_user_id,
		u.email AS installer_email,
		NULLIF(TRIM(CONCAT_WS(' ', u.last_name, u.first_name, u.middle_name)), '') AS installer_full_name,
		vpi.is_active,
		vpi.created_at,
		vpi.updated_at
	FROM vehicle_part_installations vpi
	INNER JOIN parts_catalog p ON p.id = vpi.part_id
	INNER JOIN vehicles v ON v.id = vpi.vehicle_id
	LEFT JOIN mechanic_shifts ms ON ms.id = vpi.mechanic_shift_id
	LEFT JOIN users msu ON msu.id = ms.user_id
	LEFT JOIN users u ON u.id = vpi.installed_by_user_id
`

type vehiclePartInstallationScanner interface {
	Scan(dest ...any) error
}

func scanVehiclePartInstallation(scanner vehiclePartInstallationScanner) (*models.VehiclePartInstallation, error) {
	var item models.VehiclePartInstallation
	var mechanicShiftID sql.NullInt64
	var mechanicShiftUserID sql.NullInt64
	var mechanicShiftDate sql.NullTime
	var mechanicShiftTimeFrom sql.NullTime
	var mechanicShiftTimeTo sql.NullTime
	var mechanicShiftUserEmail sql.NullString
	var mechanicShiftUserFullName sql.NullString
	var installerEmail sql.NullString
	var installerFullName sql.NullString

	if err := scanner.Scan(
		&item.ID,
		&item.PartID,
		&item.PartCatalogCode,
		&item.PartName,
		&item.PartCategory,
		&item.PartIsConsumable,
		&item.VehicleID,
		&item.VehicleStateNumber,
		&item.VehicleBrandModel,
		&mechanicShiftID,
		&mechanicShiftUserID,
		&mechanicShiftDate,
		&mechanicShiftTimeFrom,
		&mechanicShiftTimeTo,
		&mechanicShiftUserEmail,
		&mechanicShiftUserFullName,
		&item.InstalledAt,
		&item.PlannedReplacementAt,
		&item.Quantity,
		&item.InstalledByUserID,
		&installerEmail,
		&installerFullName,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	item.MechanicShiftID = nullableInt64Ptr(mechanicShiftID)
	item.MechanicShiftUserID = nullableInt64Ptr(mechanicShiftUserID)
	item.MechanicShiftDate = nullableTimePtr(mechanicShiftDate)
	item.MechanicShiftTimeFrom = nullableTimePtr(mechanicShiftTimeFrom)
	item.MechanicShiftTimeTo = nullableTimePtr(mechanicShiftTimeTo)
	item.MechanicShiftUserEmail = nullableStringPtr(mechanicShiftUserEmail)
	item.MechanicShiftUserFullName = nullableStringPtr(mechanicShiftUserFullName)
	item.InstallerEmail = nullableStringPtr(installerEmail)
	item.InstallerFullName = nullableStringPtr(installerFullName)

	return &item, nil
}

func scanVehiclePartInstallationRows(rows *sql.Rows) (*models.VehiclePartInstallation, error) {
	return scanVehiclePartInstallation(rows)
}

func getPartForUpdate(ctx context.Context, tx *sql.Tx, partID int64) (*warehousePartForInstallation, error) {
	const q = `
		SELECT id, part_id, name, quantity, category, is_consumable
		FROM parts_catalog
		WHERE id = $1
		FOR UPDATE;
	`
	var item warehousePartForInstallation
	if err := tx.QueryRowContext(ctx, q, partID).Scan(
		&item.ID,
		&item.PartID,
		&item.Name,
		&item.Quantity,
		&item.Category,
		&item.IsConsumable,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehiclePartInstallationPartNotFound
		}
		return nil, fmt.Errorf("get part for update: %w", err)
	}
	return &item, nil
}

func lockPartsForUpdate(ctx context.Context, tx *sql.Tx, ids ...int64) (map[int64]*warehousePartForInstallation, error) {
	unique := make(map[int64]struct{})
	ordered := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := unique[id]; !ok {
			unique[id] = struct{}{}
			ordered = append(ordered, id)
		}
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })

	out := make(map[int64]*warehousePartForInstallation, len(ordered))
	for _, id := range ordered {
		part, err := getPartForUpdate(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		out[id] = part
	}
	return out, nil
}

func getInstallationForUpdate(ctx context.Context, tx *sql.Tx, id int64) (*vehiclePartInstallationState, error) {
	const q = `
		SELECT id, part_id, vehicle_id, mechanic_shift_id, installed_at, planned_replacement_at, quantity, installed_by_user_id, is_active
		FROM vehicle_part_installations
		WHERE id = $1
		FOR UPDATE;
	`
	var item vehiclePartInstallationState
	var mechanicShiftID sql.NullInt64
	if err := tx.QueryRowContext(ctx, q, id).Scan(
		&item.ID,
		&item.PartID,
		&item.VehicleID,
		&mechanicShiftID,
		&item.InstalledAt,
		&item.PlannedReplacementAt,
		&item.Quantity,
		&item.InstalledByUserID,
		&item.IsActive,
	); err != nil {
		return nil, err
	}
	item.MechanicShiftID = nullableInt64Ptr(mechanicShiftID)
	return &item, nil
}

func ensureVehicleExistsTx(ctx context.Context, tx *sql.Tx, vehicleID int64) error {
	const q = `SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1);`
	var exists bool
	if err := tx.QueryRowContext(ctx, q, vehicleID).Scan(&exists); err != nil {
		return fmt.Errorf("check vehicle exists: %w", err)
	}
	if !exists {
		return ErrVehiclePartInstallationVehicleNotFound
	}
	return nil
}

func ensureInstallerExistsTx(ctx context.Context, tx *sql.Tx, userID int64) error {
	const q = `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1);`
	var exists bool
	if err := tx.QueryRowContext(ctx, q, userID).Scan(&exists); err != nil {
		return fmt.Errorf("check installer user exists: %w", err)
	}
	if !exists {
		return ErrVehiclePartInstallationUserNotFound
	}
	return nil
}

func ensureMechanicShiftExistsTx(ctx context.Context, tx *sql.Tx, mechanicShiftID int64) error {
	const q = `
		SELECT EXISTS(
			SELECT 1
			FROM mechanic_shifts
			WHERE id = $1
			  AND is_deleted = FALSE
		);
	`
	var exists bool
	if err := tx.QueryRowContext(ctx, q, mechanicShiftID).Scan(&exists); err != nil {
		return fmt.Errorf("check mechanic shift exists: %w", err)
	}
	if !exists {
		return ErrVehiclePartInstallationMechanicShiftNotFound
	}
	return nil
}

func activeInstallationExistsTx(ctx context.Context, tx *sql.Tx, partID, vehicleID, excludeID int64) (bool, error) {
	const q = `
		SELECT EXISTS(
			SELECT 1
			FROM vehicle_part_installations
			WHERE part_id = $1
			  AND vehicle_id = $2
			  AND is_active = TRUE
			  AND ($3::BIGINT = 0 OR id <> $3)
		);
	`
	var exists bool
	if err := tx.QueryRowContext(ctx, q, partID, vehicleID, excludeID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check active vehicle part installation exists: %w", err)
	}
	return exists, nil
}

func decreasePartQuantityTx(ctx context.Context, tx *sql.Tx, partID, qty int64) error {
	const q = `
		UPDATE parts_catalog
		SET quantity = quantity - $1,
			updated_at = NOW()
		WHERE id = $2;
	`
	_, err := tx.ExecContext(ctx, q, qty, partID)
	if err != nil {
		return fmt.Errorf("decrease part quantity: %w", err)
	}
	return nil
}

func increasePartQuantityTx(ctx context.Context, tx *sql.Tx, partID, qty int64) error {
	const q = `
		UPDATE parts_catalog
		SET quantity = quantity + $1,
			updated_at = NOW()
		WHERE id = $2;
	`
	_, err := tx.ExecContext(ctx, q, qty, partID)
	if err != nil {
		return fmt.Errorf("increase part quantity: %w", err)
	}
	return nil
}

func normalizeVehiclePartInstallationSortBy(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "part_id":
		return "vpi.part_id"
	case "vehicle_id":
		return "vpi.vehicle_id"
	case "mechanic_shift_id":
		return "vpi.mechanic_shift_id"
	case "quantity":
		return "vpi.quantity"
	case "installed_by_user_id":
		return "vpi.installed_by_user_id"
	case "is_active":
		return "vpi.is_active"
	case "created_at":
		return "vpi.created_at"
	case "updated_at":
		return "vpi.updated_at"
	case "planned_replacement_at":
		return "vpi.planned_replacement_at"
	default:
		return "vpi.installed_at"
	}
}

func mapVehiclePartInstallationError(err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "vehicle_part_installations_part_id_fkey"):
		return ErrVehiclePartInstallationPartNotFound
	case strings.Contains(msg, "vehicle_part_installations_vehicle_id_fkey"):
		return ErrVehiclePartInstallationVehicleNotFound
	case strings.Contains(msg, "vehicle_part_installations_installed_by_user_id_fkey"):
		return ErrVehiclePartInstallationUserNotFound
	case strings.Contains(msg, "vehicle_part_installations_mechanic_shift_id_fkey"):
		return ErrVehiclePartInstallationMechanicShiftNotFound
	default:
		return err
	}
}

func nullableInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	out := v.Int64
	return &out
}

func nullableTimePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	out := v.Time
	return &out
}

func rollbackUnlessCommitted(tx *sql.Tx) {
	_ = tx.Rollback()
}
