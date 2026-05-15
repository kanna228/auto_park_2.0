package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/user_module/models"
)

var ErrNotFound = errors.New("not found")

type DriverRepo struct {
	db     *sql.DB
	schema string
}

func NewDriverRepo(db *sql.DB, schema string) *DriverRepo {
	return &DriverRepo{db: db, schema: schema}
}

func (r *DriverRepo) table() string {
	return "drivers"
}

func (r *DriverRepo) Create(ctx context.Context, d *models.Driver) (*models.Driver, error) {
	statusID := d.StatusID
	if statusID <= 0 {
		statusID = 1
	}

	q := fmt.Sprintf(`
		INSERT INTO %s (
			iin,
			name,
			surname,
			middlename,
			phone,
			mail,
			photo_path,
			birth_date,
			license_number,
			license_category,
			experience_years,
			status_id
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id
	`, r.table())

	var id int64
	if err := r.db.QueryRowContext(ctx, q,
		d.IIN,
		d.Name,
		d.Surname,
		nullIfEmpty(d.Middlename),
		nullIfEmpty(d.Phone),
		nullIfEmpty(d.Mail),
		nullIfEmpty(d.PhotoPath),
		nullTimeValue(d.BirthDate),
		nullIfEmpty(d.LicenseNumber),
		nullIfEmpty(d.LicenseCategory),
		nullIntValue(d.ExperienceYears),
		statusID,
	).Scan(&id); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *DriverRepo) GetByID(ctx context.Context, id int64) (*models.Driver, error) {
	q := fmt.Sprintf(`
		SELECT
			d.id,
			d.iin,
			d.name,
			d.surname,
			d.middlename,
			d.phone,
			d.mail,
			d.photo_path,
			d.birth_date,
			d.license_number,
			d.license_category,
			d.experience_years,
			d.status_id,
			ds.code AS status_code,
			ds.name AS status_name,
			COALESCE(ds.description, '') AS status_description,
			ds.created_at AS status_created_at,
			ds.updated_at AS status_updated_at,
			d.created_at,
			d.updated_at
		FROM %s d
		INNER JOIN driver_statuses ds ON ds.id = d.status_id
		WHERE d.id=$1
	`, r.table())

	out, err := scanDriver(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return out, nil
}

func (r *DriverRepo) List(ctx context.Context, limit, offset int) ([]models.Driver, int64, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM %s;`, r.table())
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := fmt.Sprintf(`
		SELECT
			d.id,
			d.iin,
			d.name,
			d.surname,
			d.middlename,
			d.phone,
			d.mail,
			d.photo_path,
			d.birth_date,
			d.license_number,
			d.license_category,
			d.experience_years,
			d.status_id,
			ds.code AS status_code,
			ds.name AS status_name,
			COALESCE(ds.description, '') AS status_description,
			ds.created_at AS status_created_at,
			ds.updated_at AS status_updated_at,
			d.created_at,
			d.updated_at
		FROM %s d
		INNER JOIN driver_statuses ds ON ds.id = d.status_id
		ORDER BY d.surname ASC, d.name ASC, d.id ASC
		LIMIT $1 OFFSET $2
	`, r.table())

	rows, err := r.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	res := make([]models.Driver, 0)
	for rows.Next() {
		d, err := scanDriver(rows)
		if err != nil {
			return nil, 0, err
		}
		res = append(res, *d)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return res, total, nil
}

func (r *DriverRepo) Update(ctx context.Context, id int64, upd map[string]any) (*models.Driver, error) {
	if len(upd) == 0 {
		return r.GetByID(ctx, id)
	}

	allowed := map[string]bool{
		"iin":              true,
		"name":             true,
		"surname":          true,
		"middlename":       true,
		"phone":            true,
		"mail":             true,
		"photo_path":       true,
		"birth_date":       true,
		"license_number":   true,
		"license_category": true,
		"experience_years": true,
		"status_id":        true,
	}

	setParts := make([]string, 0, len(upd)+1)
	args := make([]any, 0, len(upd)+1)
	i := 1

	for k, v := range upd {
		if !allowed[k] {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s=$%d", k, i))
		args = append(args, v)
		i++
	}

	setParts = append(setParts, fmt.Sprintf("updated_at=$%d", i))
	args = append(args, time.Now().UTC())
	i++

	if len(setParts) == 1 {
		return r.GetByID(ctx, id)
	}

	args = append(args, id)

	q := fmt.Sprintf(`
		UPDATE %s
		SET %s
		WHERE id=$%d
	`, r.table(), strings.Join(setParts, ", "), i)

	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *DriverRepo) UpdatePhotoPath(ctx context.Context, id int64, photoPath string) (*models.Driver, error) {
	return r.Update(ctx, id, map[string]any{
		"photo_path": nullIfEmpty(photoPath),
	})
}

func (r *DriverRepo) Delete(ctx context.Context, id int64) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE id=$1`, r.table())
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *DriverRepo) UpdateStatus(ctx context.Context, id int64, statusID int64) (*models.Driver, error) {
	return r.Update(ctx, id, map[string]any{"status_id": statusID})
}

func (r *DriverRepo) StatusExists(ctx context.Context, statusID int64) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM driver_statuses WHERE id = $1);`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, statusID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check driver status exists: %w", err)
	}
	return exists, nil
}

func (r *DriverRepo) ListStatuses(ctx context.Context, limit, offset int) ([]models.DriverStatus, int64, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	const countQ = `SELECT COUNT(*) FROM driver_statuses;`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count driver statuses: %w", err)
	}

	const q = `
		SELECT id, code, name, COALESCE(description, ''), created_at, updated_at
		FROM driver_statuses
		ORDER BY id ASC
		LIMIT $1 OFFSET $2;
	`
	rows, err := r.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list driver statuses: %w", err)
	}
	defer rows.Close()

	items := make([]models.DriverStatus, 0)
	for rows.Next() {
		var item models.DriverStatus
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan driver status: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows driver statuses: %w", err)
	}

	return items, total, nil
}

func (r *DriverRepo) GetPassport(ctx context.Context, id int64) (*models.DriverPassport, error) {
	driver, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	assignedVehicles, err := r.listDriverAssignedVehicles(ctx, driver.ID)
	if err != nil {
		return nil, err
	}

	totalWorkedHours, err := r.getDriverTotalWorkedHours(ctx, id)
	if err != nil {
		return nil, err
	}

	incidentsCount, err := r.getDriverAccidentsCount(ctx, id)
	if err != nil {
		return nil, err
	}

	tripsheets, err := r.listDriverPassportTripsheets(ctx, id, 8)
	if err != nil {
		return nil, err
	}

	incidents, err := r.listDriverPassportIncidents(ctx, id, 8)
	if err != nil {
		return nil, err
	}

	return &models.DriverPassport{
		Driver:           *driver,
		Status:           driver.Status.Name,
		AssignedVehicles: assignedVehicles,
		TotalWorkedHours: totalWorkedHours,
		IncidentsCount:   incidentsCount,
		Tripsheets:       tripsheets,
		Incidents:        incidents,
	}, nil
}

func (r *DriverRepo) listDriverAssignedVehicles(ctx context.Context, driverID int64) ([]models.DriverAssignedVehicle, error) {
	const q = `
		SELECT
			v.id,
			v.board_number,
			v.state_number,
			v.brand_model,
			v.status_id,
			COALESCE(vs.name, '') AS status_name
		FROM vehicles v
		LEFT JOIN vehicle_status vs ON vs.id = v.status_id
		WHERE $1 = ANY(v.drivers_ids)
		ORDER BY v.updated_at DESC, v.id DESC;
	`

	rows, err := r.db.QueryContext(ctx, q, driverID)
	if err != nil {
		return nil, fmt.Errorf("list assigned vehicles for driver: %w", err)
	}
	defer rows.Close()

	items := make([]models.DriverAssignedVehicle, 0)
	for rows.Next() {
		var item models.DriverAssignedVehicle
		if err := rows.Scan(&item.ID, &item.BoardNumber, &item.StateNumber, &item.BrandModel, &item.StatusID, &item.StatusName); err != nil {
			return nil, fmt.Errorf("scan assigned vehicle for driver: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("assigned vehicles rows: %w", err)
	}

	return items, nil
}

func (r *DriverRepo) getDriverTotalWorkedHours(ctx context.Context, driverID int64) (float64, error) {
	const q = `
		SELECT COALESCE(SUM(EXTRACT(EPOCH FROM (end_time - start_time)) / 3600.0), 0)
		FROM tripsheets
		WHERE driver_id = $1
		  AND start_time IS NOT NULL
		  AND end_time IS NOT NULL
		  AND end_time > start_time;
	`

	var hours float64
	if err := r.db.QueryRowContext(ctx, q, driverID).Scan(&hours); err != nil {
		return 0, fmt.Errorf("get driver total worked hours: %w", err)
	}
	return hours, nil
}

func (r *DriverRepo) getDriverAccidentsCount(ctx context.Context, driverID int64) (int64, error) {
	const q = `
		SELECT COUNT(*)
		FROM incidents i
		INNER JOIN incident_types it ON it.id = i.incident_type_id
		WHERE i.driver_id = $1
		  AND LOWER(it.name) = LOWER('ДТП');
	`

	var count int64
	if err := r.db.QueryRowContext(ctx, q, driverID).Scan(&count); err != nil {
		return 0, fmt.Errorf("get driver accidents count: %w", err)
	}
	return count, nil
}

func (r *DriverRepo) listDriverPassportTripsheets(ctx context.Context, driverID int64, limit int) ([]models.DriverPassportTripsheetItem, error) {
	if limit <= 0 {
		limit = 8
	}

	const q = `
		SELECT
			t.id,
			t.tripsheet_number,
			t.tripsheet_date,
			t.vehicle_id,
			t.vehicle_brand,
			t.vehicle_plate_number,
			t.start_time,
			t.end_time,
			CASE
				WHEN t.start_time IS NOT NULL AND t.end_time IS NOT NULL AND t.end_time > t.start_time
				THEN EXTRACT(EPOCH FROM (t.end_time - t.start_time)) / 3600.0
				ELSE 0
			END AS worked_hours,
			COUNT(tt.id) AS trips_count,
			t.status_id,
			ts.name AS status_name,
			t.created_at,
			t.updated_at
		FROM tripsheets t
		LEFT JOIN tripsheet_statuses ts ON ts.id = t.status_id
		LEFT JOIN tripsheet_trips tt ON tt.tripsheet_id = t.id
		WHERE t.driver_id = $1
		GROUP BY
			t.id,
			t.tripsheet_number,
			t.tripsheet_date,
			t.vehicle_id,
			t.vehicle_brand,
			t.vehicle_plate_number,
			t.start_time,
			t.end_time,
			t.status_id,
			ts.name,
			t.created_at,
			t.updated_at
		ORDER BY t.tripsheet_date DESC, COALESCE(t.start_time, t.created_at) DESC, t.id DESC
		LIMIT $2;
	`

	rows, err := r.db.QueryContext(ctx, q, driverID, limit)
	if err != nil {
		return nil, fmt.Errorf("list driver passport tripsheets: %w", err)
	}
	defer rows.Close()

	items := make([]models.DriverPassportTripsheetItem, 0)
	for rows.Next() {
		var item models.DriverPassportTripsheetItem
		var tripDate time.Time
		var vehicleID sql.NullInt64
		var vehicleBrand sql.NullString
		var startTime sql.NullTime
		var endTime sql.NullTime
		var statusName sql.NullString

		if err := rows.Scan(
			&item.ID,
			&item.TripsheetNumber,
			&tripDate,
			&vehicleID,
			&vehicleBrand,
			&item.VehiclePlateNumber,
			&startTime,
			&endTime,
			&item.WorkedHours,
			&item.TripsCount,
			&item.StatusID,
			&statusName,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan driver passport tripsheet: %w", err)
		}

		item.TripsheetDate = tripDate.Format("2006-01-02")
		item.VehicleID = nullableInt64Ptr(vehicleID)
		item.VehicleBrand = nullableStringPtr(vehicleBrand)
		item.StatusName = nullableStringPtr(statusName)
		if startTime.Valid {
			item.StartTime = &startTime.Time
		}
		if endTime.Valid {
			item.EndTime = &endTime.Time
		}

		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("driver passport tripsheet rows: %w", err)
	}

	return items, nil
}

func (r *DriverRepo) listDriverPassportIncidents(ctx context.Context, driverID int64, limit int) ([]models.DriverPassportIncidentItem, error) {
	if limit <= 0 {
		limit = 8
	}

	const q = `
		SELECT
			i.id,
			i.incident_type_id,
			it.name AS incident_type_name,
			i.vehicle_id,
			v.state_number AS vehicle_state_number,
			i.tripsheet_id,
			i.incident_date,
			i.incident_time::text,
			i.location,
			COALESCE(i.description, ''),
			i.created_at,
			i.updated_at
		FROM incidents i
		INNER JOIN incident_types it ON it.id = i.incident_type_id
		INNER JOIN vehicles v ON v.id = i.vehicle_id
		WHERE i.driver_id = $1
		ORDER BY i.incident_date DESC, i.incident_time DESC, i.id DESC
		LIMIT $2;
	`

	rows, err := r.db.QueryContext(ctx, q, driverID, limit)
	if err != nil {
		return nil, fmt.Errorf("list driver passport incidents: %w", err)
	}
	defer rows.Close()

	items := make([]models.DriverPassportIncidentItem, 0)
	for rows.Next() {
		var item models.DriverPassportIncidentItem
		var incidentDate time.Time
		var tripsheetID sql.NullInt64

		if err := rows.Scan(
			&item.ID,
			&item.IncidentTypeID,
			&item.IncidentTypeName,
			&item.VehicleID,
			&item.VehicleStateNumber,
			&tripsheetID,
			&incidentDate,
			&item.IncidentTime,
			&item.Location,
			&item.Description,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan driver passport incident: %w", err)
		}

		item.IncidentDate = incidentDate.Format("2006-01-02")
		item.TripsheetID = nullableInt64Ptr(tripsheetID)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("driver passport incident rows: %w", err)
	}

	return items, nil
}

type driverScanner interface {
	Scan(dest ...any) error
}

func scanDriver(scanner driverScanner) (*models.Driver, error) {
	out := &models.Driver{}
	var middlename, phone, mail, photoPath sql.NullString
	var birthDate sql.NullTime
	var licenseNumber, licenseCategory sql.NullString
	var experienceYears sql.NullInt64

	if err := scanner.Scan(
		&out.ID,
		&out.IIN,
		&out.Name,
		&out.Surname,
		&middlename,
		&phone,
		&mail,
		&photoPath,
		&birthDate,
		&licenseNumber,
		&licenseCategory,
		&experienceYears,
		&out.StatusID,
		&out.Status.Code,
		&out.Status.Name,
		&out.Status.Description,
		&out.Status.CreatedAt,
		&out.Status.UpdatedAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, err
	}

	out.Status.ID = out.StatusID
	out.Middlename = middlename.String
	out.Phone = phone.String
	out.Mail = mail.String
	out.PhotoPath = photoPath.String
	if birthDate.Valid {
		out.BirthDate = &birthDate.Time
	}
	out.LicenseNumber = licenseNumber.String
	out.LicenseCategory = licenseCategory.String
	if experienceYears.Valid {
		v := int(experienceYears.Int64)
		out.ExperienceYears = &v
	}

	return out, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func nullTimeValue(v *time.Time) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullIntValue(v *int) any {
	if v == nil || *v < 0 {
		return nil
	}
	return *v
}

func nullableStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}

func nullableInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func resolveDriverStatus(assignedVehicles []models.DriverAssignedVehicle) string {
	if len(assignedVehicles) == 0 {
		return "Не закреплён"
	}
	return "Доступен"
}
