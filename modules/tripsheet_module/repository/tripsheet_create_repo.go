package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"auto_park/modules/tripsheet_module/dto"
	"auto_park/modules/tripsheet_module/models"
)

type TripsheetRepo interface {
	Create(ctx context.Context, input models.CreateTripsheetInput) (*models.Tripsheet, error)
	GetStatusIDByName(ctx context.Context, name string) (int64, error)
	GetAll(ctx context.Context, f dto.TripsheetFilter) ([]dto.TripsheetResponse, int, error)
	GetByID(ctx context.Context, id int64) (*dto.TripsheetResponse, error)
	Update(ctx context.Context, input models.UpdateTripsheetInput) (*models.Tripsheet, error)
	Delete(ctx context.Context, id int64) error
	DeleteTripsWithTripsheet(ctx context.Context, id int64) error
	ValidateDriverShiftForTripsheet(ctx context.Context, driverShiftID int64, driverID *int64, tripsheetDate time.Time) error
	ResolveDriverID(ctx context.Context, driverID *int64, driverShiftID *int64) (*int64, error)
	SetDriverStatusByCode(ctx context.Context, driverID int64, code string) error
	RefreshDriverAvailability(ctx context.Context, driverID int64) error
}

type tripsheetRepo struct {
	db     *sql.DB
	schema string
}

func NewTripsheetRepo(db *sql.DB, schema string) TripsheetRepo {
	return &tripsheetRepo{db: db, schema: schema}
}

func (r *tripsheetRepo) tripsheetsTable() string     { return "tripsheets" }
func (r *tripsheetRepo) statusesTable() string       { return "tripsheet_statuses" }
func (r *tripsheetRepo) tripsheetTripsTable() string { return "tripsheet_trips" }

func (r *tripsheetRepo) GetStatusIDByName(ctx context.Context, name string) (int64, error) {
	query := fmt.Sprintf(`
		SELECT id
		FROM %s
		WHERE name = $1
		LIMIT 1
	`, r.statusesTable())

	var id int64
	if err := r.db.QueryRowContext(ctx, query, name).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("status %q not found", name)
		}
		return 0, fmt.Errorf("get status id by name: %w", err)
	}

	return id, nil
}

func (r *tripsheetRepo) Create(ctx context.Context, input models.CreateTripsheetInput) (*models.Tripsheet, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			tripsheet_number,
			tripsheet_date,
			vehicle_id,
			vehicle_brand,
			vehicle_plate_number,
			driver_last_name,
			driver_first_name,
			driver_middle_name,
			driver_id,
			driver_shift_id,
			start_time,
			end_time,
			mileage_start,
			mileage_end,
			fuel_start,
			fuel_issued,
			fuel_consumption_theoretical,
			fuel_consumption_actual,
			status_id
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19
		)
		RETURNING
			id,
			tripsheet_number,
			tripsheet_date,
			vehicle_id,
			vehicle_brand,
			vehicle_plate_number,
			driver_last_name,
			driver_first_name,
			driver_middle_name,
			driver_id,
			driver_shift_id,
			start_time,
			end_time,
			mileage_start,
			mileage_end,
			fuel_start,
			fuel_issued,
			fuel_consumption_theoretical,
			fuel_consumption_actual,
			status_id,
			created_at,
			updated_at
	`, r.tripsheetsTable())

	var item models.Tripsheet

	err := r.db.QueryRowContext(
		ctx,
		query,
		input.TripsheetNumber,
		input.TripsheetDate,
		input.VehicleID,
		input.VehicleBrand,
		input.VehiclePlateNumber,
		input.DriverLastName,
		input.DriverFirstName,
		input.DriverMiddleName,
		input.DriverID,
		input.DriverShiftID,
		input.StartTime,
		input.EndTime,
		input.MileageStart,
		input.MileageEnd,
		input.FuelStart,
		input.FuelIssued,
		input.FuelConsumptionTheoretical,
		input.FuelConsumptionActual,
		*input.StatusID,
	).Scan(
		&item.ID,
		&item.TripsheetNumber,
		&item.TripsheetDate,
		&item.VehicleID,
		&item.VehicleBrand,
		&item.VehiclePlateNumber,
		&item.DriverLastName,
		&item.DriverFirstName,
		&item.DriverMiddleName,
		&item.DriverID,
		&item.DriverShiftID,
		&item.StartTime,
		&item.EndTime,
		&item.MileageStart,
		&item.MileageEnd,
		&item.FuelStart,
		&item.FuelIssued,
		&item.FuelConsumptionTheoretical,
		&item.FuelConsumptionActual,
		&item.StatusID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create tripsheet: %w", err)
	}

	return &item, nil
}

func (r *tripsheetRepo) ValidateDriverShiftForTripsheet(ctx context.Context, driverShiftID int64, driverID *int64, tripsheetDate time.Time) error {
	const q = `
		SELECT driver_id, shift_date
		FROM driver_shifts
		WHERE id = $1
		  AND is_deleted = FALSE;
	`

	var shiftDriverID int64
	var shiftDate time.Time
	if err := r.db.QueryRowContext(ctx, q, driverShiftID).Scan(&shiftDriverID, &shiftDate); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("driver_shift_id not found")
		}
		return fmt.Errorf("validate driver shift: %w", err)
	}

	if driverID != nil && *driverID > 0 && *driverID != shiftDriverID {
		return fmt.Errorf("driver_shift_id belongs to another driver")
	}

	if !sameDate(shiftDate, tripsheetDate) {
		return fmt.Errorf("driver_shift_id shift_date must match tripsheet_date")
	}

	return nil
}

func sameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
