package repository

import (
	"auto_park/modules/incident_module/dto"
	"auto_park/modules/incident_module/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrIncidentMechanicShiftNotFound = errors.New("mechanic shift not found or does not belong to mechanic")

type IncidentRepository interface {
	Create(ctx context.Context, input models.CreateIncidentInput) (*models.Incident, error)
	GetByID(ctx context.Context, id int64) (*models.Incident, error)
	GetAll(ctx context.Context, filter dto.IncidentListQuery) ([]models.Incident, int, error)
	Update(ctx context.Context, input models.UpdateIncidentInput) (*models.Incident, error)
	Delete(ctx context.Context, id int64) error
	ListIncidentTypes(ctx context.Context) ([]models.IncidentType, error)
}

type incidentRepo struct {
	db *sql.DB
}

func NewIncidentRepository(db *sql.DB) IncidentRepository {
	return &incidentRepo{db: db}
}

func (r *incidentRepo) createStatusToMaintenanceTx(ctx context.Context, tx *sql.Tx, vehicleID int64) error {
	const q = `
		UPDATE vehicles
		SET
			status_id = (
				SELECT id
				FROM vehicle_status
				WHERE name = 'На ТО'
				LIMIT 1
			),
			updated_at = NOW()
		WHERE id = $1;
	`

	if _, err := tx.ExecContext(ctx, q, vehicleID); err != nil {
		return fmt.Errorf("update vehicle status to maintenance: %w", err)
	}

	return nil
}

func (r *incidentRepo) ensureMechanicShiftExistsTx(ctx context.Context, tx *sql.Tx, mechanicShiftID, mechanicID int64) error {
	const q = `
		SELECT EXISTS(
			SELECT 1
			FROM mechanic_shifts
			WHERE id = $1
			  AND user_id = $2
			  AND is_deleted = FALSE
		);
	`

	var exists bool
	if err := tx.QueryRowContext(ctx, q, mechanicShiftID, mechanicID).Scan(&exists); err != nil {
		return fmt.Errorf("check mechanic shift exists: %w", err)
	}
	if !exists {
		return ErrIncidentMechanicShiftNotFound
	}

	return nil
}

func (r *incidentRepo) Create(ctx context.Context, input models.CreateIncidentInput) (*models.Incident, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx create incident: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := r.ensureMechanicShiftExistsTx(ctx, tx, input.MechanicShiftID, input.MechanicID); err != nil {
		return nil, err
	}

	const q = `
		INSERT INTO incidents (
			incident_type_id,
			vehicle_id,
			driver_id,
			mechanic_id,
			mechanic_shift_id,
			tripsheet_id,
			incident_date,
			incident_time,
			location,
			description
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id;
	`

	var id int64
	if err := tx.QueryRowContext(
		ctx,
		q,
		input.IncidentTypeID,
		input.VehicleID,
		input.DriverID,
		input.MechanicID,
		input.MechanicShiftID,
		input.TripsheetID,
		input.IncidentDate,
		input.IncidentTime,
		input.Location,
		input.Description,
	).Scan(&id); err != nil {
		return nil, fmt.Errorf("create incident: %w", err)
	}

	if err := r.createStatusToMaintenanceTx(ctx, tx, input.VehicleID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create incident: %w", err)
	}

	return r.GetByID(ctx, id)
}

const incidentSelect = `
	SELECT
		i.id,
		i.incident_type_id,
		it.name,
		i.vehicle_id,
		v.state_number,
		i.driver_id,
		TRIM(CONCAT_WS(' ', d.surname, d.name, d.middlename)) AS driver_full_name,
		d.status_id AS driver_status_id,
		ds.name AS driver_status_name,
		i.mechanic_id,
		TRIM(CONCAT_WS(' ', u.last_name, u.first_name, u.middle_name)) AS mechanic_full_name,
		i.mechanic_shift_id,

		ms.id,
		ms.user_id,
		ms.shift_date,
		ms.time_from,
		ms.time_to,
		ms.comment,
		ms.is_active,
		ms_u.email AS mechanic_shift_user_email,
		NULLIF(TRIM(CONCAT_WS(' ', ms_u.last_name, ms_u.first_name, ms_u.middle_name)), '') AS mechanic_shift_user_full_name,
		ms.created_at,
		ms.updated_at,

		i.tripsheet_id,

		t.id,
		t.tripsheet_number,
		t.tripsheet_date,
		t.vehicle_brand,
		t.vehicle_plate_number,
		t.driver_id,
		t.driver_last_name,
		t.driver_first_name,
		t.driver_middle_name,
		t.status_id,
		ts.name,
		t.start_time,
		t.end_time,
		t.mileage_start,
		t.mileage_end,
		t.fuel_start,
		t.fuel_issued,
		t.fuel_consumption_theoretical,
		t.fuel_consumption_actual,

		i.incident_date,
		i.incident_time,
		i.location,
		COALESCE(i.description, ''),
		i.created_at,
		i.updated_at
	FROM incidents i
	JOIN incident_types it ON it.id = i.incident_type_id
	JOIN vehicles v ON v.id = i.vehicle_id
	JOIN drivers d ON d.id = i.driver_id
	JOIN driver_statuses ds ON ds.id = d.status_id
	JOIN users u ON u.id = i.mechanic_id
	LEFT JOIN mechanic_shifts ms ON ms.id = i.mechanic_shift_id
	LEFT JOIN users ms_u ON ms_u.id = ms.user_id
	LEFT JOIN tripsheets t ON t.id = i.tripsheet_id
	LEFT JOIN tripsheet_statuses ts ON ts.id = t.status_id
`

type incidentScanner interface {
	Scan(dest ...any) error
}

func scanIncident(scanner incidentScanner) (*models.Incident, error) {
	var item models.Incident

	var mechanicShiftID sql.NullInt64
	var msID sql.NullInt64
	var msUserID sql.NullInt64
	var msShiftDate sql.NullTime
	var msTimeFrom sql.NullTime
	var msTimeTo sql.NullTime
	var msComment sql.NullString
	var msIsActive sql.NullBool
	var msMechanicEmail sql.NullString
	var msMechanicFullName sql.NullString
	var msCreatedAt sql.NullTime
	var msUpdatedAt sql.NullTime

	var tripsheetID sql.NullInt64
	var tID sql.NullInt64
	var tTripsheetNumber sql.NullString
	var tTripsheetDate sql.NullTime
	var tVehicleBrand sql.NullString
	var tVehiclePlateNumber sql.NullString
	var tDriverID sql.NullInt64
	var tDriverLastName sql.NullString
	var tDriverFirstName sql.NullString
	var tDriverMiddleName sql.NullString
	var tStatusID sql.NullInt64
	var tStatusName sql.NullString
	var tStartTime sql.NullTime
	var tEndTime sql.NullTime
	var tMileageStart sql.NullInt64
	var tMileageEnd sql.NullInt64
	var tFuelStart sql.NullInt64
	var tFuelIssued sql.NullInt64
	var tFuelConsumptionTheoretical sql.NullInt64
	var tFuelConsumptionActual sql.NullInt64

	if err := scanner.Scan(
		&item.ID,
		&item.IncidentTypeID,
		&item.IncidentTypeName,
		&item.VehicleID,
		&item.VehicleStateNumber,
		&item.DriverID,
		&item.DriverFullName,
		&item.DriverStatusID,
		&item.DriverStatusName,
		&item.MechanicID,
		&item.MechanicFullName,
		&mechanicShiftID,

		&msID,
		&msUserID,
		&msShiftDate,
		&msTimeFrom,
		&msTimeTo,
		&msComment,
		&msIsActive,
		&msMechanicEmail,
		&msMechanicFullName,
		&msCreatedAt,
		&msUpdatedAt,

		&tripsheetID,

		&tID,
		&tTripsheetNumber,
		&tTripsheetDate,
		&tVehicleBrand,
		&tVehiclePlateNumber,
		&tDriverID,
		&tDriverLastName,
		&tDriverFirstName,
		&tDriverMiddleName,
		&tStatusID,
		&tStatusName,
		&tStartTime,
		&tEndTime,
		&tMileageStart,
		&tMileageEnd,
		&tFuelStart,
		&tFuelIssued,
		&tFuelConsumptionTheoretical,
		&tFuelConsumptionActual,

		&item.IncidentDate,
		&item.IncidentTime,
		&item.Location,
		&item.Description,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if mechanicShiftID.Valid {
		val := mechanicShiftID.Int64
		item.MechanicShiftID = &val
	}

	if msID.Valid {
		item.MechanicShift = &models.IncidentMechanicShift{
			ID:        msID.Int64,
			UserID:    msUserID.Int64,
			ShiftDate: msShiftDate.Time,
			TimeFrom:  msTimeFrom.Time,
			IsActive:  msIsActive.Bool,
			CreatedAt: msCreatedAt.Time,
			UpdatedAt: msUpdatedAt.Time,
		}

		if msTimeTo.Valid {
			v := msTimeTo.Time
			item.MechanicShift.TimeTo = &v
		}
		if msComment.Valid {
			v := msComment.String
			item.MechanicShift.Comment = &v
		}
		if msMechanicEmail.Valid {
			v := msMechanicEmail.String
			item.MechanicShift.MechanicEmail = &v
		}
		if msMechanicFullName.Valid {
			v := msMechanicFullName.String
			item.MechanicShift.MechanicFullName = &v
		}
	}

	if tripsheetID.Valid {
		val := tripsheetID.Int64
		item.TripsheetID = &val
	}

	if tID.Valid {
		item.Tripsheet = &models.IncidentTripsheet{
			ID:                         tID.Int64,
			TripsheetNumber:            tTripsheetNumber.String,
			TripsheetDate:              tTripsheetDate.Time,
			VehiclePlateNumber:         tVehiclePlateNumber.String,
			StatusID:                   tStatusID.Int64,
			StatusName:                 tStatusName.String,
			MileageStart:               int(tMileageStart.Int64),
			MileageEnd:                 int(tMileageEnd.Int64),
			FuelStart:                  int(tFuelStart.Int64),
			FuelIssued:                 int(tFuelIssued.Int64),
			FuelConsumptionTheoretical: int(tFuelConsumptionTheoretical.Int64),
			FuelConsumptionActual:      int(tFuelConsumptionActual.Int64),
		}

		if tVehicleBrand.Valid {
			v := tVehicleBrand.String
			item.Tripsheet.VehicleBrand = &v
		}
		if tDriverID.Valid {
			v := tDriverID.Int64
			item.Tripsheet.DriverID = &v
		}
		if tDriverLastName.Valid {
			v := tDriverLastName.String
			item.Tripsheet.DriverLastName = &v
		}
		if tDriverFirstName.Valid {
			v := tDriverFirstName.String
			item.Tripsheet.DriverFirstName = &v
		}
		if tDriverMiddleName.Valid {
			v := tDriverMiddleName.String
			item.Tripsheet.DriverMiddleName = &v
		}
		if tStartTime.Valid {
			v := tStartTime.Time
			item.Tripsheet.StartTime = &v
		}
		if tEndTime.Valid {
			v := tEndTime.Time
			item.Tripsheet.EndTime = &v
		}
	}

	return &item, nil
}

func (r *incidentRepo) GetByID(ctx context.Context, id int64) (*models.Incident, error) {
	query := incidentSelect + `
		WHERE i.id = $1;
	`

	return scanIncident(r.db.QueryRowContext(ctx, query, id))
}

func (r *incidentRepo) GetAll(ctx context.Context, filter dto.IncidentListQuery) ([]models.Incident, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 12)
	idx := 1

	add := func(condition string, value any) {
		where = append(where, fmt.Sprintf(condition, idx))
		args = append(args, value)
		idx++
	}

	if filter.IncidentTypeID != nil {
		add("i.incident_type_id = $%d", *filter.IncidentTypeID)
	}
	if filter.VehicleID != nil {
		add("i.vehicle_id = $%d", *filter.VehicleID)
	}
	if filter.DriverID != nil {
		add("i.driver_id = $%d", *filter.DriverID)
	}
	if filter.MechanicID != nil {
		add("i.mechanic_id = $%d", *filter.MechanicID)
	}
	if filter.MechanicShiftID != nil {
		add("i.mechanic_shift_id = $%d", *filter.MechanicShiftID)
	}
	if filter.TripsheetID != nil {
		add("i.tripsheet_id = $%d", *filter.TripsheetID)
	}
	if s := strings.TrimSpace(filter.DateFrom); s != "" {
		add("i.incident_date >= $%d", s)
	}
	if s := strings.TrimSpace(filter.DateTo); s != "" {
		add("i.incident_date <= $%d", s)
	}

	whereSQL := strings.Join(where, " AND ")
	countQuery := `
		SELECT COUNT(*)
		FROM incidents i
		JOIN incident_types it ON it.id = i.incident_type_id
		JOIN vehicles v ON v.id = i.vehicle_id
		JOIN drivers d ON d.id = i.driver_id
		JOIN users u ON u.id = i.mechanic_id
		LEFT JOIN mechanic_shifts ms ON ms.id = i.mechanic_shift_id
		LEFT JOIN users ms_u ON ms_u.id = ms.user_id
		LEFT JOIN tripsheets t ON t.id = i.tripsheet_id
		LEFT JOIN tripsheet_statuses ts ON ts.id = t.status_id
		WHERE ` + whereSQL + `;`

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count incidents: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	sortBy := normalizeIncidentSortBy(filter.SortBy)
	order := normalizeIncidentOrder(filter.Order)

	query := incidentSelect + `
		WHERE ` + whereSQL + fmt.Sprintf(`
		ORDER BY %s %s, i.id DESC
		LIMIT $%d OFFSET $%d;
	`, sortBy, order, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]models.Incident, 0)
	for rows.Next() {
		item, err := scanIncident(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func normalizeIncidentSortBy(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "id":
		return "i.id"
	case "incident_type_id":
		return "i.incident_type_id"
	case "vehicle_id":
		return "i.vehicle_id"
	case "driver_id":
		return "i.driver_id"
	case "mechanic_id":
		return "i.mechanic_id"
	case "mechanic_shift_id":
		return "i.mechanic_shift_id"
	case "tripsheet_id":
		return "i.tripsheet_id"
	case "created_at":
		return "i.created_at"
	case "updated_at":
		return "i.updated_at"
	default:
		return "i.incident_date"
	}
}

func normalizeIncidentOrder(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "asc") {
		return "ASC"
	}
	return "DESC"
}

func (r *incidentRepo) Update(ctx context.Context, input models.UpdateIncidentInput) (*models.Incident, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx update incident: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := r.ensureMechanicShiftExistsTx(ctx, tx, input.MechanicShiftID, input.MechanicID); err != nil {
		return nil, err
	}

	const q = `
		UPDATE incidents
		SET
			incident_type_id = $1,
			vehicle_id = $2,
			driver_id = $3,
			mechanic_id = $4,
			mechanic_shift_id = $5,
			tripsheet_id = $6,
			incident_date = $7,
			incident_time = $8,
			location = $9,
			description = $10,
			updated_at = NOW()
		WHERE id = $11;
	`

	res, err := tx.ExecContext(
		ctx,
		q,
		input.IncidentTypeID,
		input.VehicleID,
		input.DriverID,
		input.MechanicID,
		input.MechanicShiftID,
		input.TripsheetID,
		input.IncidentDate,
		input.IncidentTime,
		input.Location,
		input.Description,
		input.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update incident: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("update incident rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	if err := r.createStatusToMaintenanceTx(ctx, tx, input.VehicleID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update incident: %w", err)
	}

	return r.GetByID(ctx, input.ID)
}

func (r *incidentRepo) Delete(ctx context.Context, id int64) error {
	const q = `
		DELETE FROM incidents
		WHERE id = $1;
	`

	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete incident: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete incident rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *incidentRepo) ListIncidentTypes(ctx context.Context) ([]models.IncidentType, error) {
	const q = `
		SELECT id, name, created_at, updated_at
		FROM incident_types
		ORDER BY id ASC;
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list incident types: %w", err)
	}
	defer rows.Close()

	items := make([]models.IncidentType, 0)
	for rows.Next() {
		var item models.IncidentType
		if err := rows.Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan incident type: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("incident type rows: %w", err)
	}

	return items, nil
}
