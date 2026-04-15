package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"auto_park/modules/fuel_module/models"
)

type FuelReportRepository interface {
	BuildDriverReport(ctx context.Context, driverID int64, dateFrom, dateTo *time.Time) (*models.FuelReportData, error)
	BuildVehicleReport(ctx context.Context, vehicleID int64, dateFrom, dateTo *time.Time) (*models.FuelReportData, error)
	BuildTripsheetReport(ctx context.Context, tripsheetID int64, dateFrom, dateTo *time.Time) (*models.FuelReportData, error)
}

type fuelReportRepository struct {
	db *sql.DB
}

func NewFuelReportRepository(db *sql.DB) FuelReportRepository {
	return &fuelReportRepository{db: db}
}

type reportBaseMeta struct {
	Title    string
	Subtitle string
}

type rowAccumulator struct {
	rows    []models.FuelReportRow
	daily   map[string]*models.FuelReportDailyStat
	summary models.FuelReportSummary
}

func newAccumulator() *rowAccumulator {
	return &rowAccumulator{daily: make(map[string]*models.FuelReportDailyStat)}
}

func (r *fuelReportRepository) BuildDriverReport(ctx context.Context, driverID int64, dateFrom, dateTo *time.Time) (*models.FuelReportData, error) {
	meta, err := r.getDriverMeta(ctx, driverID)
	if err != nil {
		return nil, err
	}

	rows, err := r.fetchRows(ctx, reportQueryScopeDriver, driverID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	data := buildReportData(models.FuelReportEntityDriver, driverID, meta, dateFrom, dateTo, rows)
	return data, nil
}

func (r *fuelReportRepository) BuildVehicleReport(ctx context.Context, vehicleID int64, dateFrom, dateTo *time.Time) (*models.FuelReportData, error) {
	meta, err := r.getVehicleMeta(ctx, vehicleID)
	if err != nil {
		return nil, err
	}

	rows, err := r.fetchRows(ctx, reportQueryScopeVehicle, vehicleID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	data := buildReportData(models.FuelReportEntityVehicle, vehicleID, meta, dateFrom, dateTo, rows)
	return data, nil
}

func (r *fuelReportRepository) BuildTripsheetReport(ctx context.Context, tripsheetID int64, dateFrom, dateTo *time.Time) (*models.FuelReportData, error) {
	meta, err := r.getTripsheetMeta(ctx, tripsheetID)
	if err != nil {
		return nil, err
	}

	rows, err := r.fetchRows(ctx, reportQueryScopeTripsheet, tripsheetID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	data := buildReportData(models.FuelReportEntityTripsheet, tripsheetID, meta, dateFrom, dateTo, rows)
	return data, nil
}

type reportQueryScope string

const (
	reportQueryScopeDriver    reportQueryScope = "driver"
	reportQueryScopeVehicle   reportQueryScope = "vehicle"
	reportQueryScopeTripsheet reportQueryScope = "tripsheet"
)

func (r *fuelReportRepository) fetchRows(ctx context.Context, scope reportQueryScope, entityID int64, dateFrom, dateTo *time.Time) ([]models.FuelReportRow, error) {
	var conditions []string
	args := []interface{}{}
	argPos := 1

	switch scope {
	case reportQueryScopeDriver:
		conditions = append(conditions, fmt.Sprintf("t.driver_id = $%d", argPos))
		args = append(args, entityID)
		argPos++
	case reportQueryScopeVehicle:
		conditions = append(conditions, fmt.Sprintf("fr.vehicle_id = $%d", argPos))
		args = append(args, entityID)
		argPos++
	case reportQueryScopeTripsheet:
		conditions = append(conditions, fmt.Sprintf("fr.tripsheet_id = $%d", argPos))
		args = append(args, entityID)
		argPos++
	}

	if dateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("fr.date >= $%d", argPos))
		args = append(args, *dateFrom)
		argPos++
	}
	if dateTo != nil {
		conditions = append(conditions, fmt.Sprintf("fr.date <= $%d", argPos))
		args = append(args, *dateTo)
		argPos++
	}

	query := `
SELECT
	fr.id,
	fr.tripsheet_id,
	COALESCE(t.tripsheet_number, ''),
	t.tripsheet_date,
	fr.vehicle_id,
	COALESCE(v.brand_model, '') AS vehicle_brand_model,
	COALESCE(v.state_number, '') AS vehicle_state_number,
	COALESCE(v.board_number, '') AS vehicle_board_number,
	COALESCE(t.driver_id, 0),
	COALESCE(d.surname, ''),
	COALESCE(d.name, ''),
	COALESCE(d.middlename, ''),
	fr.date,
	fr.time,
	COALESCE(fr.location, ''),
	fr.fuel_amount,
	COALESCE(t.mileage_start, 0),
	COALESCE(t.mileage_end, 0),
	COALESCE(t.fuel_start, 0),
	COALESCE(t.fuel_issued, 0),
	COALESCE(t.fuel_consumption_actual, 0),
	COALESCE(t.fuel_consumption_theoretical, 0),
	COALESCE(tt.distance_km, 0)
FROM fuel_refills fr
LEFT JOIN tripsheets t ON t.id = fr.tripsheet_id
LEFT JOIN vehicles v ON v.id = fr.vehicle_id
LEFT JOIN drivers d ON d.id = t.driver_id
LEFT JOIN (
	SELECT tripsheet_id, COALESCE(SUM(distance_passed), 0) AS distance_km
	FROM tripsheet_trips
	GROUP BY tripsheet_id
) tt ON tt.tripsheet_id = fr.tripsheet_id
`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY fr.date ASC, fr.time ASC, fr.id ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query fuel report rows: %w", err)
	}
	defer rows.Close()

	var out []models.FuelReportRow
	for rows.Next() {
		var item models.FuelReportRow
		var tripsheetDate sql.NullTime
		var refillTime time.Time
		var driverID int64
		var driverSurname, driverName, driverMiddle string
		var vehicleBrandModel, vehicleStateNumber, vehicleBoardNumber string

		if err := rows.Scan(
			&item.RefillID,
			&item.TripsheetID,
			&item.TripsheetNumber,
			&tripsheetDate,
			&item.VehicleID,
			&vehicleBrandModel,
			&vehicleStateNumber,
			&vehicleBoardNumber,
			&driverID,
			&driverSurname,
			&driverName,
			&driverMiddle,
			&item.RefillDate,
			&refillTime,
			&item.Location,
			&item.FuelAmount,
			&item.MileageStart,
			&item.MileageEnd,
			&item.FuelStart,
			&item.FuelIssued,
			&item.FuelActual,
			&item.FuelTheoretical,
			&item.DistancePassedKM,
		); err != nil {
			return nil, fmt.Errorf("scan fuel report row: %w", err)
		}

		item.RefillTime = refillTime.Format("15:04:05")
		item.DriverID = driverID
		item.DriverName = joinNonEmpty([]string{driverSurname, driverName, driverMiddle}, " ")
		item.VehicleLabel = strings.TrimSpace(joinNonEmpty([]string{vehicleBrandModel, vehicleStateNumber, vehicleBoardNumber}, " | "))
		if tripsheetDate.Valid {
			item.TripsheetDate = &tripsheetDate.Time
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fuel report rows: %w", err)
	}
	return out, nil
}

func (r *fuelReportRepository) getDriverMeta(ctx context.Context, driverID int64) (reportBaseMeta, error) {
	query := `SELECT id, surname, name, middlename, iin, COALESCE(phone, '') FROM drivers WHERE id = $1`
	var id int64
	var surname, name, middlename, iin, phone string
	if err := r.db.QueryRowContext(ctx, query, driverID).Scan(&id, &surname, &name, &middlename, &iin, &phone); err != nil {
		if err == sql.ErrNoRows {
			return reportBaseMeta{}, sql.ErrNoRows
		}
		return reportBaseMeta{}, fmt.Errorf("get driver meta: %w", err)
	}
	return reportBaseMeta{
		Title:    fmt.Sprintf("Fuel report by driver #%d", id),
		Subtitle: strings.TrimSpace(fmt.Sprintf("Driver: %s | IIN: %s | Phone: %s", joinNonEmpty([]string{surname, name, middlename}, " "), iin, emptyFallback(phone, "-"))),
	}, nil
}

func (r *fuelReportRepository) getVehicleMeta(ctx context.Context, vehicleID int64) (reportBaseMeta, error) {
	query := `SELECT id, board_number, state_number, brand_model, vin, COALESCE(technical_passport_number, '') FROM vehicles WHERE id = $1`
	var id int64
	var boardNumber, stateNumber, brandModel, vin, techPass string
	if err := r.db.QueryRowContext(ctx, query, vehicleID).Scan(&id, &boardNumber, &stateNumber, &brandModel, &vin, &techPass); err != nil {
		if err == sql.ErrNoRows {
			return reportBaseMeta{}, sql.ErrNoRows
		}
		return reportBaseMeta{}, fmt.Errorf("get vehicle meta: %w", err)
	}
	return reportBaseMeta{
		Title:    fmt.Sprintf("Fuel report by vehicle #%d", id),
		Subtitle: fmt.Sprintf("Vehicle: %s | State No: %s | Board No: %s | VIN: %s | Tech passport: %s", brandModel, stateNumber, boardNumber, vin, emptyFallback(techPass, "-")),
	}, nil
}

func (r *fuelReportRepository) getTripsheetMeta(ctx context.Context, tripsheetID int64) (reportBaseMeta, error) {
	query := `
SELECT
	t.id,
	COALESCE(t.tripsheet_number, ''),
	t.tripsheet_date,
	COALESCE(t.vehicle_brand, ''),
	COALESCE(t.vehicle_plate_number, ''),
	COALESCE(t.driver_last_name, ''),
	COALESCE(t.driver_first_name, ''),
	COALESCE(t.driver_middle_name, ''),
	COALESCE(t.mileage_start, 0),
	COALESCE(t.mileage_end, 0),
	COALESCE(t.fuel_start, 0),
	COALESCE(t.fuel_issued, 0),
	COALESCE(t.fuel_consumption_actual, 0),
	COALESCE(t.fuel_consumption_theoretical, 0)
FROM tripsheets t
WHERE t.id = $1`
	var id int64
	var number, tVehicleBrand, tVehiclePlate, dLast, dFirst, dMiddle string
	var tDate time.Time
	var mileageStart, mileageEnd, fuelStart, fuelIssued, fuelActual, fuelTheoretical int
	if err := r.db.QueryRowContext(ctx, query, tripsheetID).Scan(
		&id, &number, &tDate, &tVehicleBrand, &tVehiclePlate, &dLast, &dFirst, &dMiddle,
		&mileageStart, &mileageEnd, &fuelStart, &fuelIssued, &fuelActual, &fuelTheoretical,
	); err != nil {
		if err == sql.ErrNoRows {
			return reportBaseMeta{}, sql.ErrNoRows
		}
		return reportBaseMeta{}, fmt.Errorf("get tripsheet meta: %w", err)
	}
	return reportBaseMeta{
		Title: fmt.Sprintf("Fuel report by tripsheet #%d", id),
		Subtitle: fmt.Sprintf(
			"Tripsheet: %s from %s | Vehicle: %s / %s | Driver: %s | Mileage: %d-%d | Fuel start: %d | Fuel issued: %d | Actual: %d | Theoretical: %d",
			number, tDate.Format("2006-01-02"), tVehicleBrand, tVehiclePlate,
			joinNonEmpty([]string{dLast, dFirst, dMiddle}, " "), mileageStart, mileageEnd, fuelStart, fuelIssued, fuelActual, fuelTheoretical,
		),
	}, nil
}

func buildReportData(entityType models.FuelReportEntityType, entityID int64, meta reportBaseMeta, dateFrom, dateTo *time.Time, rows []models.FuelReportRow) *models.FuelReportData {
	acc := newAccumulator()
	for _, row := range rows {
		acc.add(row)
	}

	dailyStats := make([]models.FuelReportDailyStat, 0, len(acc.daily))
	for _, item := range acc.daily {
		sort.Slice(item.TripsheetIDs, func(i, j int) bool { return item.TripsheetIDs[i] < item.TripsheetIDs[j] })
		dailyStats = append(dailyStats, *item)
	}
	sort.Slice(dailyStats, func(i, j int) bool { return dailyStats[i].Date.Before(dailyStats[j].Date) })

	return &models.FuelReportData{
		Meta: models.FuelReportMeta{
			EntityType:  entityType,
			EntityID:    entityID,
			Title:       meta.Title,
			Subtitle:    meta.Subtitle,
			PeriodLabel: buildPeriodLabel(dateFrom, dateTo),
			GeneratedAt: time.Now(),
		},
		Summary:    acc.summary,
		Rows:       rows,
		DailyStats: dailyStats,
	}
}

func (a *rowAccumulator) add(row models.FuelReportRow) {
	a.rows = append(a.rows, row)
	a.summary.RefillCount++
	a.summary.TotalFuelAmount += row.FuelAmount
	a.summary.TotalDistanceKM += row.DistancePassedKM
	if a.summary.RefillCount > 0 {
		a.summary.AverageFuelPerOp = a.summary.TotalFuelAmount / float64(a.summary.RefillCount)
	}
	if a.summary.FirstRefillDate == nil || row.RefillDate.Before(*a.summary.FirstRefillDate) {
		copyDate := row.RefillDate
		a.summary.FirstRefillDate = &copyDate
	}
	if a.summary.LastRefillDate == nil || row.RefillDate.After(*a.summary.LastRefillDate) {
		copyDate := row.RefillDate
		a.summary.LastRefillDate = &copyDate
	}

	key := row.RefillDate.Format("2006-01-02")
	item, ok := a.daily[key]
	if !ok {
		dateOnly, _ := time.Parse("2006-01-02", key)
		item = &models.FuelReportDailyStat{Date: dateOnly}
		a.daily[key] = item
	}
	item.RefillCount++
	item.TotalFuel += row.FuelAmount
	if row.TripsheetID > 0 && !containsInt64(item.TripsheetIDs, row.TripsheetID) {
		item.TripsheetIDs = append(item.TripsheetIDs, row.TripsheetID)
	}
}

func buildPeriodLabel(dateFrom, dateTo *time.Time) string {
	switch {
	case dateFrom != nil && dateTo != nil:
		return fmt.Sprintf("%s - %s", dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	case dateFrom != nil:
		return fmt.Sprintf("from %s", dateFrom.Format("2006-01-02"))
	case dateTo != nil:
		return fmt.Sprintf("until %s", dateTo.Format("2006-01-02"))
	default:
		return "all time"
	}
}

func containsInt64(items []int64, target int64) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func joinNonEmpty(items []string, sep string) string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return strings.Join(out, sep)
}

func emptyFallback(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
