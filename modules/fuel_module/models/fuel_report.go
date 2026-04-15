package models

import "time"

type FuelReportEntityType string

const (
	FuelReportEntityDriver    FuelReportEntityType = "driver"
	FuelReportEntityVehicle   FuelReportEntityType = "vehicle"
	FuelReportEntityTripsheet FuelReportEntityType = "tripsheet"
)

type FuelReportFile struct {
	FileName    string
	ContentType string
	Data        []byte
}

type FuelReportMeta struct {
	EntityType  FuelReportEntityType
	EntityID    int64
	Title       string
	Subtitle    string
	PeriodLabel string
	GeneratedAt time.Time
}

type FuelReportSummary struct {
	RefillCount      int
	TotalFuelAmount  float64
	FirstRefillDate  *time.Time
	LastRefillDate   *time.Time
	TotalDistanceKM  int
	AverageFuelPerOp float64
}

type FuelReportRow struct {
	RefillID         int64
	TripsheetID      int64
	TripsheetNumber  string
	TripsheetDate    *time.Time
	VehicleID        int64
	VehicleLabel     string
	DriverID         int64
	DriverName       string
	RefillDate       time.Time
	RefillTime       string
	Location         string
	FuelAmount       float64
	MileageStart     int
	MileageEnd       int
	DistancePassedKM int
	FuelStart        int
	FuelIssued       int
	FuelActual       int
	FuelTheoretical  int
}

type FuelReportDailyStat struct {
	Date         time.Time
	RefillCount  int
	TotalFuel    float64
	TripsheetIDs []int64
}

type FuelReportData struct {
	Meta       FuelReportMeta
	Summary    FuelReportSummary
	Rows       []FuelReportRow
	DailyStats []FuelReportDailyStat
}
