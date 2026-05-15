package repository

import (
	"auto_park/modules/vehicle_module/models"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

func (r *vehicleRepo) GetByID(ctx context.Context, id int64) (*models.Vehicle, error) {
	const q = `
		SELECT
			v.id,
			v.board_number,
			v.technical_passport_number,
			v.state_number,
			v.vin,
			v.brand_model,
			v.manufacture_year,
			v.received_date,
			v.empty_weight_kg,
			v.max_weight_kg,
			v.engine_volume_cc,
			v.insurance_policy_number,
			v.insurance_expiry_date,
			v.mileage,
			v.current_fuel,
			v.status_id,
			vs.name AS status_name,
			v.drivers_ids,
			v.photo_path,
			v.created_at,
			v.updated_at
		FROM vehicles v
		JOIN vehicle_status vs ON vs.id = v.status_id
		WHERE v.id = $1;
	`

	var v models.Vehicle
	var drivers pq.Int64Array
	var photoPath sql.NullString

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&v.ID,
		&v.BoardNumber,
		&v.TechnicalPassportNumber,
		&v.StateNumber,
		&v.VIN,
		&v.BrandModel,
		&v.ManufactureYear,
		&v.ReceivedDate,
		&v.EmptyWeightKG,
		&v.MaxWeightKG,
		&v.EngineVolumeCC,
		&v.InsurancePolicyNumber,
		&v.InsuranceExpiryDate,
		&v.Mileage,
		&v.CurrentFuel,
		&v.StatusID,
		&v.StatusName,
		&drivers,
		&photoPath,
		&v.CreatedAt,
		&v.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get vehicle by id: %w", err)
	}

	v.DriversIDs = []int64(drivers)
	v.PhotoPath = photoPath.String
	return &v, nil
}

type ListVehiclesParams struct {
	BoardNumber string
	StateNumber string
	VIN         string
	BrandModel  string

	StatusID *int64

	ManufactureYearFrom *int
	ManufactureYearTo   *int

	DriverID *int64

	Limit  int
	Offset int

	SortBy string
	Order  string
}

func (r *vehicleRepo) List(ctx context.Context, q ListVehiclesParams) ([]models.Vehicle, int64, error) {
	where := make([]string, 0, 10)
	args := make([]any, 0, 14)

	add := func(cond string, val any) {
		args = append(args, val)
		where = append(where, fmt.Sprintf(cond, len(args)))
	}

	if strings.TrimSpace(q.BoardNumber) != "" {
		add("v.board_number ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.BoardNumber))
	}
	if strings.TrimSpace(q.StateNumber) != "" {
		add("v.state_number ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.StateNumber))
	}
	if strings.TrimSpace(q.VIN) != "" {
		add("v.vin ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.VIN))
	}
	if strings.TrimSpace(q.BrandModel) != "" {
		add("v.brand_model ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.BrandModel))
	}

	if q.StatusID != nil {
		add("v.status_id = $%d", *q.StatusID)
	}

	if q.ManufactureYearFrom != nil {
		add("v.manufacture_year >= $%d", *q.ManufactureYearFrom)
	}
	if q.ManufactureYearTo != nil {
		add("v.manufacture_year <= $%d", *q.ManufactureYearTo)
	}

	if q.DriverID != nil {
		add("$%d = ANY(v.drivers_ids)", *q.DriverID)
	}

	whereSQL := "TRUE"
	if len(where) > 0 {
		whereSQL = strings.Join(where, " AND ")
	}

	sortCol := "v.id"
	switch q.SortBy {
	case "id":
		sortCol = "v.id"
	case "board_number":
		sortCol = "v.board_number"
	case "state_number":
		sortCol = "v.state_number"
	case "manufacture_year":
		sortCol = "v.manufacture_year"
	case "mileage":
		sortCol = "v.mileage"
	case "created_at":
		sortCol = "v.created_at"
	case "status_id":
		sortCol = "v.status_id"
	case "status_name":
		sortCol = "vs.name"
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

	countSQL := `
		SELECT COUNT(*)
		FROM vehicles v
		JOIN vehicle_status vs ON vs.id = v.status_id
		WHERE ` + whereSQL + `;`

	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list vehicles count: %w", err)
	}

	argsData := append([]any{}, args...)
	argsData = append(argsData, limit, offset)

	dataSQL := fmt.Sprintf(`
		SELECT
			v.id,
			v.board_number,
			v.technical_passport_number,
			v.state_number,
			v.vin,
			v.brand_model,
			v.manufacture_year,
			v.received_date,
			v.empty_weight_kg,
			v.max_weight_kg,
			v.engine_volume_cc,
			v.insurance_policy_number,
			v.insurance_expiry_date,
			v.mileage,
			v.current_fuel,
			v.status_id,
			vs.name AS status_name,
			v.drivers_ids,
			v.photo_path,
			v.created_at,
			v.updated_at
		FROM vehicles v
		JOIN vehicle_status vs ON vs.id = v.status_id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortCol, order, len(args)+1, len(args)+2)

	rows, err := r.db.QueryContext(ctx, dataSQL, argsData...)
	if err != nil {
		return nil, 0, fmt.Errorf("list vehicles query: %w", err)
	}
	defer rows.Close()

	items := make([]models.Vehicle, 0, limit)
	for rows.Next() {
		var v models.Vehicle
		var drivers pq.Int64Array
		var photoPath sql.NullString

		if err := rows.Scan(
			&v.ID,
			&v.BoardNumber,
			&v.TechnicalPassportNumber,
			&v.StateNumber,
			&v.VIN,
			&v.BrandModel,
			&v.ManufactureYear,
			&v.ReceivedDate,
			&v.EmptyWeightKG,
			&v.MaxWeightKG,
			&v.EngineVolumeCC,
			&v.InsurancePolicyNumber,
			&v.InsuranceExpiryDate,
			&v.Mileage,
			&v.CurrentFuel,
			&v.StatusID,
			&v.StatusName,
			&drivers,
			&photoPath,
			&v.CreatedAt,
			&v.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("list vehicles scan: %w", err)
		}

		v.DriversIDs = []int64(drivers)
		v.PhotoPath = photoPath.String
		items = append(items, v)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list vehicles rows: %w", err)
	}

	return items, total, nil
}

type VehicleIncidentHistoryRow struct {
	ID               int64
	IncidentTypeID   int64
	IncidentTypeName string

	VehicleID int64

	TripsheetID     *int64
	TripsheetNumber *string

	IncidentDate time.Time
	IncidentTime string
	Location     string
	Description  string

	DriverID         int64
	DriverIIN        string
	DriverFirstName  string
	DriverLastName   string
	DriverMiddleName *string
	DriverPhone      *string
	DriverEmail      *string
	DriverStatusID   int64
	DriverStatusName string

	MechanicID         int64
	MechanicIIN        string
	MechanicFirstName  string
	MechanicLastName   string
	MechanicMiddleName *string
	MechanicPhone      *string
	MechanicEmail      string
	MechanicRoleID     int64
	MechanicRoleName   string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *vehicleRepo) ListIncidentsByVehicleID(ctx context.Context, vehicleID int64) ([]VehicleIncidentHistoryRow, error) {
	const q = `
		SELECT
			i.id,
			i.incident_type_id,
			it.name AS incident_type_name,
			i.vehicle_id,
			i.tripsheet_id,
			t.tripsheet_number,
			i.incident_date,
			TO_CHAR(i.incident_time, 'HH24:MI:SS') AS incident_time,
			i.location,
			COALESCE(i.description, '') AS description,

			d.id,
			d.iin,
			d.name,
			d.surname,
			d.middlename,
			d.phone,
			d.mail,
			d.status_id,
			ds.name AS driver_status_name,

			u.id,
			u.iin,
			u.first_name,
			u.last_name,
			u.middle_name,
			u.phone,
			u.email,
			r.id,
			r.name,

			i.created_at,
			i.updated_at
		FROM incidents i
		JOIN incident_types it ON it.id = i.incident_type_id
		JOIN drivers d ON d.id = i.driver_id
		JOIN driver_statuses ds ON ds.id = d.status_id
		JOIN users u ON u.id = i.mechanic_id
		JOIN roles r ON r.id = u.role_id
		LEFT JOIN tripsheets t ON t.id = i.tripsheet_id
		WHERE i.vehicle_id = $1
		ORDER BY i.incident_date DESC, i.incident_time DESC, i.id DESC
		LIMIT 8;
	`

	rows, err := r.db.QueryContext(ctx, q, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("list incidents by vehicle id: %w", err)
	}
	defer rows.Close()

	items := make([]VehicleIncidentHistoryRow, 0)
	for rows.Next() {
		var item VehicleIncidentHistoryRow

		var tripsheetID sql.NullInt64
		var tripsheetNumber sql.NullString

		var driverMiddleName sql.NullString
		var driverPhone sql.NullString
		var driverEmail sql.NullString

		var mechanicMiddleName sql.NullString
		var mechanicPhone sql.NullString

		if err := rows.Scan(
			&item.ID,
			&item.IncidentTypeID,
			&item.IncidentTypeName,
			&item.VehicleID,
			&tripsheetID,
			&tripsheetNumber,
			&item.IncidentDate,
			&item.IncidentTime,
			&item.Location,
			&item.Description,

			&item.DriverID,
			&item.DriverIIN,
			&item.DriverFirstName,
			&item.DriverLastName,
			&driverMiddleName,
			&driverPhone,
			&driverEmail,
			&item.DriverStatusID,
			&item.DriverStatusName,

			&item.MechanicID,
			&item.MechanicIIN,
			&item.MechanicFirstName,
			&item.MechanicLastName,
			&mechanicMiddleName,
			&mechanicPhone,
			&item.MechanicEmail,
			&item.MechanicRoleID,
			&item.MechanicRoleName,

			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan incident by vehicle id: %w", err)
		}

		if tripsheetID.Valid {
			v := tripsheetID.Int64
			item.TripsheetID = &v
		}
		if tripsheetNumber.Valid {
			v := tripsheetNumber.String
			item.TripsheetNumber = &v
		}

		if driverMiddleName.Valid {
			v := driverMiddleName.String
			item.DriverMiddleName = &v
		}
		if driverPhone.Valid {
			v := driverPhone.String
			item.DriverPhone = &v
		}
		if driverEmail.Valid {
			v := driverEmail.String
			item.DriverEmail = &v
		}

		if mechanicMiddleName.Valid {
			v := mechanicMiddleName.String
			item.MechanicMiddleName = &v
		}
		if mechanicPhone.Valid {
			v := mechanicPhone.String
			item.MechanicPhone = &v
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows incidents by vehicle id: %w", err)
	}

	return items, nil
}

type VehicleDriverRow struct {
	ID         int64
	IIN        string
	FirstName  string
	LastName   string
	MiddleName *string
	Phone      *string
	Email      *string
	StatusID   int64
	StatusName string
}

func (r *vehicleRepo) ListDriversByIDs(ctx context.Context, ids []int64) ([]VehicleDriverRow, error) {
	if len(ids) == 0 {
		return []VehicleDriverRow{}, nil
	}

	const q = `
		SELECT
			d.id,
			d.iin,
			d.name,
			d.surname,
			d.middlename,
			d.phone,
			d.mail,
			d.status_id,
			ds.name AS status_name
		FROM drivers d
		INNER JOIN driver_statuses ds ON ds.id = d.status_id
		WHERE d.id = ANY($1)
		ORDER BY d.id ASC;
	`

	rows, err := r.db.QueryContext(ctx, q, pq.Int64Array(ids))
	if err != nil {
		return nil, fmt.Errorf("list drivers by ids: %w", err)
	}
	defer rows.Close()

	items := make([]VehicleDriverRow, 0, len(ids))
	for rows.Next() {
		var item VehicleDriverRow
		var middleName sql.NullString
		var phone sql.NullString
		var email sql.NullString

		if err := rows.Scan(
			&item.ID,
			&item.IIN,
			&item.FirstName,
			&item.LastName,
			&middleName,
			&phone,
			&email,
			&item.StatusID,
			&item.StatusName,
		); err != nil {
			return nil, fmt.Errorf("scan drivers by ids: %w", err)
		}

		if middleName.Valid {
			v := middleName.String
			item.MiddleName = &v
		}
		if phone.Valid {
			v := phone.String
			item.Phone = &v
		}
		if email.Valid {
			v := email.String
			item.Email = &v
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows drivers by ids: %w", err)
	}

	return items, nil
}

type VehicleInstalledPartHistoryRow struct {
	ID                   int64
	PartID               int64
	CatalogPartID        string
	PartName             string
	PartCategory         string
	IsConsumable         bool
	VehicleID            int64
	InstalledAt          time.Time
	PlannedReplacementAt *time.Time
	Quantity             int64
	UnitPrice            float64
	TotalPrice           float64
	InstalledByUserID    int64
	InstallerEmail       *string
	InstallerFullName    *string
	IsActive             bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (r *vehicleRepo) ListInstalledPartsByVehicleID(ctx context.Context, vehicleID int64) ([]VehicleInstalledPartHistoryRow, error) {
	const q = `
		SELECT
			vpi.id,
			vpi.part_id,
			p.part_id AS catalog_part_id,
			p.name AS part_name,
			p.category AS part_category,
			p.is_consumable,
			vpi.vehicle_id,
			vpi.installed_at,
			vpi.planned_replacement_at,
			vpi.quantity,
			vpi.unit_price,
			vpi.total_price,
			vpi.installed_by_user_id,
			u.email AS installer_email,
			NULLIF(TRIM(CONCAT_WS(' ', u.last_name, u.first_name, u.middle_name)), '') AS installer_full_name,
			vpi.is_active,
			vpi.created_at,
			vpi.updated_at
		FROM vehicle_part_installations vpi
		JOIN parts_catalog p ON p.id = vpi.part_id
		LEFT JOIN users u ON u.id = vpi.installed_by_user_id
		WHERE vpi.vehicle_id = $1
		ORDER BY vpi.installed_at DESC, vpi.id DESC
		LIMIT 8;
	`

	rows, err := r.db.QueryContext(ctx, q, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("list installed parts by vehicle id: %w", err)
	}
	defer rows.Close()

	items := make([]VehicleInstalledPartHistoryRow, 0)

	for rows.Next() {
		var item VehicleInstalledPartHistoryRow

		var plannedReplacementAt sql.NullTime
		var installerEmail sql.NullString
		var installerFullName sql.NullString

		if err := rows.Scan(
			&item.ID,
			&item.PartID,
			&item.CatalogPartID,
			&item.PartName,
			&item.PartCategory,
			&item.IsConsumable,
			&item.VehicleID,
			&item.InstalledAt,
			&plannedReplacementAt,
			&item.Quantity,
			&item.UnitPrice,
			&item.TotalPrice,
			&item.InstalledByUserID,
			&installerEmail,
			&installerFullName,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan installed part by vehicle id: %w", err)
		}

		if plannedReplacementAt.Valid {
			v := plannedReplacementAt.Time
			item.PlannedReplacementAt = &v
		}

		if installerEmail.Valid {
			v := installerEmail.String
			item.InstallerEmail = &v
		}

		if installerFullName.Valid {
			v := installerFullName.String
			item.InstallerFullName = &v
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows installed parts by vehicle id: %w", err)
	}

	return items, nil
}

type VehicleTripsheetTripHistoryRow struct {
	ID               int64
	TripsheetID      int64
	RouteDescription string
	StartTime        *time.Time
	EndTime          *time.Time
	DistancePassed   int
	StatusID         int64
	StatusName       string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type VehicleTripsheetHistoryRow struct {
	ID                         int64
	TripsheetNumber            string
	TripsheetDate              time.Time
	VehicleID                  *int64
	VehicleBrand               *string
	VehiclePlateNumber         string
	DriverID                   *int64
	DriverShiftID              *int64
	DriverIIN                  *string
	DriverFirstName            *string
	DriverLastName             *string
	DriverMiddleName           *string
	DriverPhone                *string
	DriverEmail                *string
	DriverStatusID             int64
	DriverStatusName           *string
	DriverSnapshotLastName     *string
	DriverSnapshotFirstName    *string
	DriverSnapshotMiddleName   *string
	StartTime                  *time.Time
	EndTime                    *time.Time
	MileageStart               int
	MileageEnd                 int
	FuelStart                  int
	FuelIssued                 int
	FuelConsumptionTheoretical int
	FuelConsumptionActual      int
	StatusID                   int64
	StatusName                 string
	Trips                      []VehicleTripsheetTripHistoryRow
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

func (r *vehicleRepo) ListTripsheetsByVehicleID(ctx context.Context, vehicleID int64) ([]VehicleTripsheetHistoryRow, error) {
	const q = `
		SELECT
			t.id,
			t.tripsheet_number,
			t.tripsheet_date,
			t.vehicle_id,
			t.vehicle_brand,
			t.vehicle_plate_number,
			t.driver_id,
			t.driver_shift_id,
			d.iin AS driver_iin,
			d.name AS driver_first_name,
			d.surname AS driver_last_name,
			d.middlename AS driver_middle_name,
			d.phone AS driver_phone,
			d.mail AS driver_email,
			d.status_id AS driver_status_id,
			dst.name AS driver_status_name,
			t.driver_last_name AS driver_snapshot_last_name,
			t.driver_first_name AS driver_snapshot_first_name,
			t.driver_middle_name AS driver_snapshot_middle_name,
			t.start_time,
			t.end_time,
			t.mileage_start,
			t.mileage_end,
			t.fuel_start,
			t.fuel_issued,
			t.fuel_consumption_theoretical,
			t.fuel_consumption_actual,
			t.status_id,
			COALESCE(ts.name, '') AS status_name,
			t.created_at,
			t.updated_at
		FROM tripsheets t
		LEFT JOIN tripsheet_statuses ts ON ts.id = t.status_id
		LEFT JOIN drivers d ON d.id = t.driver_id
		LEFT JOIN driver_statuses dst ON dst.id = d.status_id
		WHERE t.vehicle_id = $1
		   OR t.vehicle_plate_number = (
			   SELECT state_number
			   FROM vehicles
			   WHERE id = $1
		   )
		ORDER BY t.tripsheet_date DESC, t.id DESC
		LIMIT 8;
	`

	rows, err := r.db.QueryContext(ctx, q, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("list tripsheets by vehicle id: %w", err)
	}
	defer rows.Close()

	items := make([]VehicleTripsheetHistoryRow, 0)
	tripsheetIDs := make([]int64, 0)
	indexByID := make(map[int64]int)

	for rows.Next() {
		var item VehicleTripsheetHistoryRow

		var vehicleIDValue sql.NullInt64
		var vehicleBrand sql.NullString
		var driverID sql.NullInt64
		var driverShiftID sql.NullInt64
		var driverIIN sql.NullString
		var driverFirstName sql.NullString
		var driverLastName sql.NullString
		var driverMiddleName sql.NullString
		var driverPhone sql.NullString
		var driverEmail sql.NullString
		var driverStatusID sql.NullInt64
		var driverStatusName sql.NullString
		var driverSnapshotLastName sql.NullString
		var driverSnapshotFirstName sql.NullString
		var driverSnapshotMiddleName sql.NullString
		var startTime sql.NullTime
		var endTime sql.NullTime

		if err := rows.Scan(
			&item.ID,
			&item.TripsheetNumber,
			&item.TripsheetDate,
			&vehicleIDValue,
			&vehicleBrand,
			&item.VehiclePlateNumber,
			&driverID,
			&driverShiftID,
			&driverIIN,
			&driverFirstName,
			&driverLastName,
			&driverMiddleName,
			&driverPhone,
			&driverEmail,
			&driverStatusID,
			&driverStatusName,
			&driverSnapshotLastName,
			&driverSnapshotFirstName,
			&driverSnapshotMiddleName,
			&startTime,
			&endTime,
			&item.MileageStart,
			&item.MileageEnd,
			&item.FuelStart,
			&item.FuelIssued,
			&item.FuelConsumptionTheoretical,
			&item.FuelConsumptionActual,
			&item.StatusID,
			&item.StatusName,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan tripsheet by vehicle id: %w", err)
		}

		item.VehicleID = nullableInt64Ptr(vehicleIDValue)
		item.VehicleBrand = nullableStringPtrVehicle(vehicleBrand)
		item.DriverID = nullableInt64Ptr(driverID)
		item.DriverShiftID = nullableInt64Ptr(driverShiftID)
		item.DriverIIN = nullableStringPtrVehicle(driverIIN)
		item.DriverFirstName = nullableStringPtrVehicle(driverFirstName)
		item.DriverLastName = nullableStringPtrVehicle(driverLastName)
		item.DriverMiddleName = nullableStringPtrVehicle(driverMiddleName)
		item.DriverPhone = nullableStringPtrVehicle(driverPhone)
		item.DriverEmail = nullableStringPtrVehicle(driverEmail)
		if driverStatusID.Valid {
			item.DriverStatusID = driverStatusID.Int64
		}
		item.DriverStatusName = nullableStringPtrVehicle(driverStatusName)
		item.DriverSnapshotLastName = nullableStringPtrVehicle(driverSnapshotLastName)
		item.DriverSnapshotFirstName = nullableStringPtrVehicle(driverSnapshotFirstName)
		item.DriverSnapshotMiddleName = nullableStringPtrVehicle(driverSnapshotMiddleName)
		item.StartTime = nullableTimePtr(startTime)
		item.EndTime = nullableTimePtr(endTime)
		item.Trips = []VehicleTripsheetTripHistoryRow{}

		indexByID[item.ID] = len(items)
		tripsheetIDs = append(tripsheetIDs, item.ID)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows tripsheets by vehicle id: %w", err)
	}

	if len(tripsheetIDs) == 0 {
		return items, nil
	}

	trips, err := r.listTripsheetTripsByTripsheetIDs(ctx, tripsheetIDs)
	if err != nil {
		return nil, err
	}

	for _, trip := range trips {
		idx, ok := indexByID[trip.TripsheetID]
		if !ok {
			continue
		}
		items[idx].Trips = append(items[idx].Trips, trip)
	}

	return items, nil
}

func (r *vehicleRepo) listTripsheetTripsByTripsheetIDs(ctx context.Context, tripsheetIDs []int64) ([]VehicleTripsheetTripHistoryRow, error) {
	const q = `
		SELECT
			tt.id,
			tt.tripsheet_id,
			tt.route_description,
			tt.start_time,
			tt.end_time,
			tt.distance_passed,
			tt.status_id,
			COALESCE(ts.name, '') AS status_name,
			tt.created_at,
			tt.updated_at
		FROM tripsheet_trips tt
		LEFT JOIN tripsheet_statuses ts ON ts.id = tt.status_id
		WHERE tt.tripsheet_id = ANY($1)
		ORDER BY tt.tripsheet_id ASC, tt.start_time ASC NULLS LAST, tt.id ASC;
	`

	rows, err := r.db.QueryContext(ctx, q, pq.Int64Array(tripsheetIDs))
	if err != nil {
		return nil, fmt.Errorf("list tripsheet trips by tripsheet ids: %w", err)
	}
	defer rows.Close()

	items := make([]VehicleTripsheetTripHistoryRow, 0)
	for rows.Next() {
		var item VehicleTripsheetTripHistoryRow
		var startTime sql.NullTime
		var endTime sql.NullTime

		if err := rows.Scan(
			&item.ID,
			&item.TripsheetID,
			&item.RouteDescription,
			&startTime,
			&endTime,
			&item.DistancePassed,
			&item.StatusID,
			&item.StatusName,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan tripsheet trip by tripsheet ids: %w", err)
		}

		item.StartTime = nullableTimePtr(startTime)
		item.EndTime = nullableTimePtr(endTime)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows tripsheet trips by tripsheet ids: %w", err)
	}

	return items, nil
}

type VehicleServiceHistoryRow struct {
	ID              int64
	TypeID          int64
	TypeName        string
	TypeDescription *string
	PartID          int64
	PartName        string
	PartDescription *string
	VehicleID       int64
	ServiceDate     time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (r *vehicleRepo) ListVehicleServicesByVehicleID(ctx context.Context, vehicleID int64) ([]VehicleServiceHistoryRow, error) {
	const q = `
		SELECT
			vs.id,
			vs.type_id,
			st.name AS type_name,
			st.description AS type_description,
			vs.part_id,
			pc.name AS part_name,
			pc.description AS part_description,
			vs.vehicle_id,
			vs.service_date,
			vs.created_at,
			vs.updated_at
		FROM vehicle_services vs
		INNER JOIN service_types st ON st.id = vs.type_id
		INNER JOIN parts_collection pc ON pc.id = vs.part_id
		WHERE vs.vehicle_id = $1
		ORDER BY vs.service_date DESC, vs.id DESC
		LIMIT 8;
	`

	rows, err := r.db.QueryContext(ctx, q, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("list vehicle services by vehicle id: %w", err)
	}
	defer rows.Close()

	items := make([]VehicleServiceHistoryRow, 0)
	for rows.Next() {
		var item VehicleServiceHistoryRow
		var typeDescription sql.NullString
		var partDescription sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.TypeID,
			&item.TypeName,
			&typeDescription,
			&item.PartID,
			&item.PartName,
			&partDescription,
			&item.VehicleID,
			&item.ServiceDate,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vehicle service by vehicle id: %w", err)
		}
		item.TypeDescription = nullableStringPtrVehicle(typeDescription)
		item.PartDescription = nullableStringPtrVehicle(partDescription)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows vehicle services by vehicle id: %w", err)
	}

	return items, nil
}

func nullableInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	value := v.Int64
	return &value
}

func nullableStringPtrVehicle(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	value := v.String
	return &value
}

func nullableTimePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	value := v.Time
	return &value
}
