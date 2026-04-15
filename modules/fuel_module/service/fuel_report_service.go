package service

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/fuel_module/dto"
	"auto_park/modules/fuel_module/models"
	"auto_park/modules/fuel_module/repository"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

type FuelReportService interface {
	DownloadDriverReport(ctx context.Context, driverID int64, filter dto.FuelReportFilter) (*models.FuelReportFile, error)
	DownloadVehicleReport(ctx context.Context, vehicleID int64, filter dto.FuelReportFilter) (*models.FuelReportFile, error)
	DownloadTripsheetReport(ctx context.Context, tripsheetID int64, filter dto.FuelReportFilter) (*models.FuelReportFile, error)
}

type fuelReportService struct {
	repo repository.FuelReportRepository
}

func NewFuelReportService(repo repository.FuelReportRepository) FuelReportService {
	return &fuelReportService{repo: repo}
}

func (s *fuelReportService) DownloadDriverReport(ctx context.Context, driverID int64, filter dto.FuelReportFilter) (*models.FuelReportFile, error) {
	if driverID <= 0 {
		return nil, fmt.Errorf("invalid driver_id")
	}
	dateFrom, dateTo, format, err := validateReportFilter(filter)
	if err != nil {
		return nil, err
	}
	data, err := s.repo.BuildDriverReport(ctx, driverID, dateFrom, dateTo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return s.renderReport(data, format)
}

func (s *fuelReportService) DownloadVehicleReport(ctx context.Context, vehicleID int64, filter dto.FuelReportFilter) (*models.FuelReportFile, error) {
	if vehicleID <= 0 {
		return nil, fmt.Errorf("invalid vehicle_id")
	}
	dateFrom, dateTo, format, err := validateReportFilter(filter)
	if err != nil {
		return nil, err
	}
	data, err := s.repo.BuildVehicleReport(ctx, vehicleID, dateFrom, dateTo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return s.renderReport(data, format)
}

func (s *fuelReportService) DownloadTripsheetReport(ctx context.Context, tripsheetID int64, filter dto.FuelReportFilter) (*models.FuelReportFile, error) {
	if tripsheetID <= 0 {
		return nil, fmt.Errorf("invalid tripsheet_id")
	}
	dateFrom, dateTo, format, err := validateReportFilter(filter)
	if err != nil {
		return nil, err
	}
	data, err := s.repo.BuildTripsheetReport(ctx, tripsheetID, dateFrom, dateTo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return s.renderReport(data, format)
}

func validateReportFilter(filter dto.FuelReportFilter) (*time.Time, *time.Time, string, error) {
	format := strings.ToLower(strings.TrimSpace(filter.Format))
	if format == "" {
		format = "pdf"
	}
	if format != "pdf" && format != "xlsx" {
		return nil, nil, "", fmt.Errorf("format must be pdf or xlsx")
	}

	var dateFrom *time.Time
	if filter.DateFrom != nil && strings.TrimSpace(*filter.DateFrom) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*filter.DateFrom))
		if err != nil {
			return nil, nil, "", fmt.Errorf("invalid date_from format, expected YYYY-MM-DD")
		}
		dateFrom = &parsed
	}
	var dateTo *time.Time
	if filter.DateTo != nil && strings.TrimSpace(*filter.DateTo) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*filter.DateTo))
		if err != nil {
			return nil, nil, "", fmt.Errorf("invalid date_to format, expected YYYY-MM-DD")
		}
		dateTo = &parsed
	}
	if dateFrom != nil && dateTo != nil && dateFrom.After(*dateTo) {
		return nil, nil, "", fmt.Errorf("date_from must be less than or equal to date_to")
	}
	return dateFrom, dateTo, format, nil
}

func (s *fuelReportService) renderReport(data *models.FuelReportData, format string) (*models.FuelReportFile, error) {
	safeType := string(data.Meta.EntityType)
	safePeriod := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(data.Meta.PeriodLabel, " ", "_"), ":", "-"), "/", "-")
	baseName := fmt.Sprintf("fuel_report_%s_%d_%s", safeType, data.Meta.EntityID, safePeriod)

	switch format {
	case "xlsx":
		bytesData, err := buildFuelReportExcel(data)
		if err != nil {
			return nil, err
		}
		return &models.FuelReportFile{
			FileName:    baseName + ".xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			Data:        bytesData,
		}, nil
	default:
		bytesData, err := buildFuelReportPDF(data)
		if err != nil {
			return nil, err
		}
		return &models.FuelReportFile{
			FileName:    baseName + ".pdf",
			ContentType: "application/pdf",
			Data:        bytesData,
		}, nil
	}
}

func buildFuelReportPDF(data *models.FuelReportData) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 10)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 8, data.Meta.Title, "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 5, data.Meta.Subtitle, "", "L", false)
	pdf.CellFormat(0, 6, "Period: "+data.Meta.PeriodLabel, "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, "Generated at: "+data.Meta.GeneratedAt.Format("2006-01-02 15:04:05"), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 7, "Summary", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(65, 6, fmt.Sprintf("Refill count: %d", data.Summary.RefillCount), "1", 0, "L", false, 0, "")
	pdf.CellFormat(65, 6, fmt.Sprintf("Total fuel amount: %.2f", data.Summary.TotalFuelAmount), "1", 0, "L", false, 0, "")
	pdf.CellFormat(65, 6, fmt.Sprintf("Avg fuel per refill: %.2f", data.Summary.AverageFuelPerOp), "1", 0, "L", false, 0, "")
	pdf.CellFormat(65, 6, fmt.Sprintf("Total distance (km): %d", data.Summary.TotalDistanceKM), "1", 1, "L", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 7, "Daily totals", "", 1, "L", false, 0, "")
	renderPDFDailyTable(pdf, data)
	pdf.Ln(3)

	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 7, "Refill details", "", 1, "L", false, 0, "")
	renderPDFRowsTable(pdf, data)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("build fuel report pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func renderPDFDailyTable(pdf *gofpdf.Fpdf, data *models.FuelReportData) {
	widths := []float64{35, 35, 45, 150}
	headers := []string{"Date", "Refills", "Total fuel", "Tripsheets"}
	renderPDFHeader(pdf, widths, headers)
	pdf.SetFont("Arial", "", 9)
	if len(data.DailyStats) == 0 {
		renderPDFSingleRow(pdf, widths, "No data for selected period")
		return
	}
	for _, item := range data.DailyStats {
		row := []string{
			item.Date.Format("2006-01-02"),
			fmt.Sprintf("%d", item.RefillCount),
			fmt.Sprintf("%.2f", item.TotalFuel),
			joinInt64(item.TripsheetIDs),
		}
		renderPDFRow(pdf, widths, row)
	}
}

func renderPDFRowsTable(pdf *gofpdf.Fpdf, data *models.FuelReportData) {
	widths := []float64{12, 20, 18, 18, 28, 42, 40, 40, 22, 18}
	headers := []string{"ID", "Date", "Time", "Trip ID", "Trip No", "Vehicle", "Driver", "Location", "Fuel", "Dist"}
	renderPDFHeader(pdf, widths, headers)
	pdf.SetFont("Arial", "", 8)
	if len(data.Rows) == 0 {
		renderPDFSingleRow(pdf, widths, "No refills found")
		return
	}
	for _, item := range data.Rows {
		row := []string{
			fmt.Sprintf("%d", item.RefillID),
			item.RefillDate.Format("2006-01-02"),
			item.RefillTime,
			fmt.Sprintf("%d", item.TripsheetID),
			emptyDash(item.TripsheetNumber),
			trimTo(item.VehicleLabel, 30),
			trimTo(item.DriverName, 28),
			trimTo(item.Location, 28),
			fmt.Sprintf("%.2f", item.FuelAmount),
			fmt.Sprintf("%d", item.DistancePassedKM),
		}
		renderPDFRow(pdf, widths, row)
	}
}

func renderPDFHeader(pdf *gofpdf.Fpdf, widths []float64, headers []string) {
	pdf.SetFont("Arial", "B", 9)
	for i, h := range headers {
		pdf.CellFormat(widths[i], 7, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)
}

func renderPDFRow(pdf *gofpdf.Fpdf, widths []float64, cells []string) {
	for i, cell := range cells {
		pdf.CellFormat(widths[i], 6, cell, "1", 0, "L", false, 0, "")
	}
	pdf.Ln(-1)
}

func renderPDFSingleRow(pdf *gofpdf.Fpdf, widths []float64, text string) {
	total := 0.0
	for _, w := range widths {
		total += w
	}
	pdf.CellFormat(total, 6, text, "1", 1, "L", false, 0, "")
}

func buildFuelReportExcel(data *models.FuelReportData) ([]byte, error) {
	f := excelize.NewFile()
	summarySheet := "Summary"
	dailySheet := "Daily totals"
	refillsSheet := "Refills"
	f.SetSheetName("Sheet1", summarySheet)
	_, _ = f.NewSheet(dailySheet)
	_, _ = f.NewSheet(refillsSheet)

	writeSummarySheet(f, summarySheet, data)
	writeDailySheet(f, dailySheet, data)
	writeRefillsSheet(f, refillsSheet, data)
	f.SetActiveSheet(0)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("build fuel report excel: %w", err)
	}
	return buf.Bytes(), nil
}

func writeSummarySheet(f *excelize.File, sheet string, data *models.FuelReportData) {
	rows := [][]interface{}{
		{"Title", data.Meta.Title},
		{"Subtitle", data.Meta.Subtitle},
		{"Period", data.Meta.PeriodLabel},
		{"Generated at", data.Meta.GeneratedAt.Format("2006-01-02 15:04:05")},
		{},
		{"Refill count", data.Summary.RefillCount},
		{"Total fuel amount", data.Summary.TotalFuelAmount},
		{"Average fuel per refill", data.Summary.AverageFuelPerOp},
		{"Total distance (km)", data.Summary.TotalDistanceKM},
	}
	for i, row := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		_ = f.SetSheetRow(sheet, cell, &row)
	}
	_ = f.SetColWidth(sheet, "A", "A", 28)
	_ = f.SetColWidth(sheet, "B", "B", 120)
}

func writeDailySheet(f *excelize.File, sheet string, data *models.FuelReportData) {
	headers := []interface{}{"Date", "Refills", "Total fuel", "Tripsheets"}
	_ = f.SetSheetRow(sheet, "A1", &headers)
	for i, item := range data.DailyStats {
		row := []interface{}{
			item.Date.Format("2006-01-02"),
			item.RefillCount,
			item.TotalFuel,
			joinInt64(item.TripsheetIDs),
		}
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		_ = f.SetSheetRow(sheet, cell, &row)
	}
	_ = f.SetColWidth(sheet, "A", "D", 26)
}

func writeRefillsSheet(f *excelize.File, sheet string, data *models.FuelReportData) {
	headers := []interface{}{"Refill ID", "Date", "Time", "Tripsheet ID", "Tripsheet number", "Tripsheet date", "Vehicle ID", "Vehicle", "Driver ID", "Driver", "Location", "Fuel amount", "Distance km", "Mileage start", "Mileage end", "Fuel start", "Fuel issued", "Fuel actual", "Fuel theoretical"}
	_ = f.SetSheetRow(sheet, "A1", &headers)
	for i, item := range data.Rows {
		tripDate := ""
		if item.TripsheetDate != nil {
			tripDate = item.TripsheetDate.Format("2006-01-02")
		}
		row := []interface{}{
			item.RefillID,
			item.RefillDate.Format("2006-01-02"),
			item.RefillTime,
			item.TripsheetID,
			item.TripsheetNumber,
			tripDate,
			item.VehicleID,
			item.VehicleLabel,
			item.DriverID,
			item.DriverName,
			item.Location,
			item.FuelAmount,
			item.DistancePassedKM,
			item.MileageStart,
			item.MileageEnd,
			item.FuelStart,
			item.FuelIssued,
			item.FuelActual,
			item.FuelTheoretical,
		}
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		_ = f.SetSheetRow(sheet, cell, &row)
	}
	_ = f.SetColWidth(sheet, "A", "S", 20)
}

func joinInt64(items []int64) string {
	if len(items) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("%d", item))
	}
	return strings.Join(parts, ", ")
}

func emptyDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}

func trimTo(v string, max int) string {
	v = strings.TrimSpace(v)
	if len(v) <= max {
		return emptyDash(v)
	}
	if max <= 3 {
		return v[:max]
	}
	return v[:max-3] + "..."
}
