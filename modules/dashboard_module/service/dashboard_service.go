package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"auto_park/internal/excelreport"
	"auto_park/modules/dashboard_module/dto"
	"auto_park/modules/dashboard_module/repository"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

type DashboardService interface {
	GetStats(ctx context.Context) (*dto.DashboardStatsResponse, error)
	GetAnalytics(ctx context.Context, period string) (*dto.AnalyticsDashboardResponse, error)
	ExportStatsExcel(ctx context.Context) ([]byte, string, error)
	ExportAnalyticsExcel(ctx context.Context, period string) ([]byte, string, error)
	ExportStatsPDF(ctx context.Context) ([]byte, string, error)
}

type dashboardService struct {
	repo repository.DashboardRepository
}

func NewDashboardService(repo repository.DashboardRepository) DashboardService {
	return &dashboardService{repo: repo}
}

func (s *dashboardService) GetStats(ctx context.Context) (*dto.DashboardStatsResponse, error) {
	return s.repo.GetStats(ctx)
}

func (s *dashboardService) GetAnalytics(ctx context.Context, period string) (*dto.AnalyticsDashboardResponse, error) {
	return s.repo.GetAnalytics(ctx, period)
}

func (s *dashboardService) ExportStatsExcel(ctx context.Context) ([]byte, string, error) {
	stats, err := s.repo.GetExportStats(ctx)
	if err != nil {
		return nil, "", err
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	const sheetKPI = "KPI"
	f.SetSheetName("Sheet1", sheetKPI)

	styles := excelreport.NewStyles(f)
	writeKPIExcelSheet(f, sheetKPI, stats, styles)
	writeMonthlyExcelSheet(f, stats, styles)
	writeFleetExcelSheet(f, stats, styles)
	writeStockExcelSheet(f, stats, styles)
	f.SetActiveSheet(0)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", fmt.Errorf("write dashboard excel: %w", err)
	}

	return buf.Bytes(), dashboardExportFilename("xlsx"), nil
}

func (s *dashboardService) ExportAnalyticsExcel(ctx context.Context, period string) ([]byte, string, error) {
	analytics, err := s.repo.GetAnalytics(ctx, period)
	if err != nil {
		return nil, "", err
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	styles := excelreport.NewStyles(f)
	const overview = "Analytics overview"
	f.SetSheetName("Sheet1", overview)

	writeAnalyticsOverviewSheet(f, overview, analytics, styles)
	writeAnalyticsMonthlySheet(f, analytics, styles)
	writeAnalyticsCriticalPartsSheet(f, analytics, styles)
	f.SetActiveSheet(0)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", fmt.Errorf("write analytics excel: %w", err)
	}

	return buf.Bytes(), analyticsExportFilename(analytics.Period, "xlsx"), nil
}

func (s *dashboardService) ExportStatsPDF(ctx context.Context) ([]byte, string, error) {
	stats, err := s.repo.GetExportStats(ctx)
	if err != nil {
		return nil, "", err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	fontName := setupDashboardPDFFont(pdf)
	pdf.SetTitle("Auto Park Dashboard Report", false)
	pdf.SetAuthor("auto_park", false)
	pdf.SetMargins(12, 12, 12)
	pdf.SetAutoPageBreak(true, 14)
	pdf.AddPage()
	pdf.SetFont(fontName, "B", 16)
	pdf.CellFormat(0, 9, "Auto Park Dashboard Report", "", 1, "L", false, 0, "")
	pdf.SetFont(fontName, "", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf("Week: %s - %s", stats.Week.DateFrom, stats.Week.DateTo), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, "Generated at: "+time.Now().Format("2006-01-02 15:04:05"), "", 1, "L", false, 0, "")
	pdf.Ln(4)

	writePDFSectionTitle(pdf, fontName, "KPI cards")
	writePDFSimpleRows(pdf, fontName, [][]string{
		{"Fuel expenses current", money(stats.Cards.FuelExpenses.Current)},
		{"Fuel expenses previous", money(stats.Cards.FuelExpenses.Previous)},
		{"Fuel change", percent(stats.Cards.FuelExpenses.ChangePercent)},
		{"Repair expenses current", money(stats.Cards.RepairExpenses.Current)},
		{"Repair expenses previous", money(stats.Cards.RepairExpenses.Previous)},
		{"Repair change", percent(stats.Cards.RepairExpenses.ChangePercent)},
		{"Issued parts current", strconv.FormatInt(stats.Cards.IssuedParts.Current, 10)},
		{"Pending part requests", strconv.FormatInt(stats.Cards.IssuedParts.PendingRequests, 10)},
		{"Warehouse stock value", money(stats.Cards.WarehouseStock.Current)},
		{"Critical stock positions", strconv.FormatInt(stats.Cards.WarehouseStock.CriticalCount, 10)},
	})

	writePDFSectionTitle(pdf, fontName, "Monthly expenses")
	writePDFTable(pdf, fontName, []float64{36, 42, 42, 42}, []string{"Month", "Fuel", "Repair", "Total"}, monthlyRows(stats))

	writePDFSectionTitle(pdf, fontName, "Fleet status")
	writePDFTable(pdf, fontName, []float64{25, 95, 35}, []string{"ID", "Status", "Count"}, fleetRows(stats))

	writePDFSectionTitle(pdf, fontName, "Warehouse stock - all positions")
	writePDFTable(pdf, fontName, []float64{16, 38, 49, 23, 22, 24, 24}, []string{"ID", "Part ID", "Name", "Qty", "Price", "Total", "Consumable"}, stockRows(stats))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", fmt.Errorf("write dashboard pdf: %w", err)
	}
	return buf.Bytes(), dashboardExportFilename("pdf"), nil
}

func writeKPIExcelSheet(f *excelize.File, sheet string, stats *dto.DashboardStatsResponse, styles excelreport.Styles) {
	rows := [][]any{
		{"Metric", "Current", "Previous", "Change %", "Trend"},
		{"Fuel expenses", stats.Cards.FuelExpenses.Current, stats.Cards.FuelExpenses.Previous, stats.Cards.FuelExpenses.ChangePercent, stats.Cards.FuelExpenses.Trend},
		{"Repair expenses", stats.Cards.RepairExpenses.Current, stats.Cards.RepairExpenses.Previous, stats.Cards.RepairExpenses.ChangePercent, stats.Cards.RepairExpenses.Trend},
		{"Issued parts", stats.Cards.IssuedParts.Current, stats.Cards.IssuedParts.Previous, stats.Cards.IssuedParts.ChangePercent, stats.Cards.IssuedParts.Trend},
		{"Pending part requests", stats.Cards.IssuedParts.PendingRequests, "", "", ""},
		{"Warehouse stock value", stats.Cards.WarehouseStock.Current, "", "", ""},
		{"Critical stock positions", stats.Cards.WarehouseStock.CriticalCount, "", "", ""},
		{"Week from", stats.Week.DateFrom, "", "", ""},
		{"Week to", stats.Week.DateTo, "", "", ""},
	}
	excelreport.WriteRows(f, sheet, 1, rows)
	excelreport.ApplyTable(f, sheet, 1, len(rows), 5, styles, true)
	excelreport.FreezeBelow(f, sheet, 1)
	excelreport.SetWidths(f, sheet, []float64{28, 18, 18, 14, 14})
	_ = f.SetCellStyle(sheet, "B2", "C3", styles.Money)
	_ = f.SetCellStyle(sheet, "B6", "B6", styles.Money)
	_ = f.SetCellStyle(sheet, "B7", "B7", styles.Integer)
}

func writeMonthlyExcelSheet(f *excelize.File, stats *dto.DashboardStatsResponse, styles excelreport.Styles) {
	const sheet = "Monthly expenses"
	_, _ = f.NewSheet(sheet)
	rows := [][]any{{"Month", "Fuel expenses", "Repair expenses", "Total expenses"}}
	for _, item := range stats.MonthlyExpenses {
		rows = append(rows, []any{item.Month, item.FuelExpenses, item.RepairExpenses, item.TotalExpenses})
	}
	excelreport.WriteRows(f, sheet, 1, rows)
	excelreport.ApplyTable(f, sheet, 1, len(rows), 4, styles, true)
	excelreport.FreezeBelow(f, sheet, 1)
	excelreport.SetWidths(f, sheet, []float64{16, 18, 18, 18})
	if len(rows) > 1 {
		_ = f.SetCellStyle(sheet, "B2", fmt.Sprintf("D%d", len(rows)), styles.Money)
	}
}

func writeFleetExcelSheet(f *excelize.File, stats *dto.DashboardStatsResponse, styles excelreport.Styles) {
	const sheet = "Fleet status"
	_, _ = f.NewSheet(sheet)
	rows := [][]any{{"Status ID", "Status name", "Count"}}
	for _, item := range stats.FleetStatus.Items {
		rows = append(rows, []any{item.StatusID, item.StatusName, item.Count})
	}
	rows = append(rows, []any{"", "Total", stats.FleetStatus.Total})
	excelreport.WriteRows(f, sheet, 1, rows)
	excelreport.ApplyTable(f, sheet, 1, len(rows), 3, styles, true)
	excelreport.FreezeBelow(f, sheet, 1)
	excelreport.SetWidths(f, sheet, []float64{14, 34, 14})
	if len(rows) > 1 {
		_ = f.SetCellStyle(sheet, "C2", fmt.Sprintf("C%d", len(rows)), styles.Integer)
	}
}

func writeStockExcelSheet(f *excelize.File, stats *dto.DashboardStatsResponse, styles excelreport.Styles) {
	const sheet = "Warehouse stock"
	_, _ = f.NewSheet(sheet)
	rows := [][]any{{"ID", "Part ID", "Name", "Category", "Quantity", "Price", "Total value", "Is consumable"}}
	for _, item := range stats.WarehouseStock.Items {
		rows = append(rows, []any{item.ID, item.PartID, item.Name, item.Category, item.Quantity, item.Price, item.TotalValue, item.IsConsumable})
	}
	excelreport.WriteRows(f, sheet, 1, rows)
	excelreport.ApplyTable(f, sheet, 1, len(rows), 8, styles, true)
	excelreport.FreezeBelow(f, sheet, 1)
	excelreport.SetWidths(f, sheet, []float64{10, 20, 36, 22, 14, 16, 18, 16})
	if len(rows) > 1 {
		_ = f.SetCellStyle(sheet, "E2", fmt.Sprintf("E%d", len(rows)), styles.Integer)
		_ = f.SetCellStyle(sheet, "F2", fmt.Sprintf("G%d", len(rows)), styles.Money)
	}
}

func writeAnalyticsOverviewSheet(f *excelize.File, sheet string, analytics *dto.AnalyticsDashboardResponse, styles excelreport.Styles) {
	_ = f.SetCellValue(sheet, "A1", "Analytics report")
	_ = f.SetCellStyle(sheet, "A1", "A1", styles.Title)
	_ = f.SetCellValue(sheet, "A2", "Period")
	_ = f.SetCellValue(sheet, "B2", analytics.Period)
	_ = f.SetCellValue(sheet, "A3", "Generated at")
	_ = f.SetCellValue(sheet, "B3", time.Now().Format("2006-01-02 15:04:05"))

	kpiRows := [][]any{
		{"Metric", "Value"},
		{"Fuel cost", analytics.KPI.FuelCost},
		{"Repair cost", analytics.KPI.RepairCost},
		{"Parts issued", analytics.KPI.PartsIssued},
		{"Warehouse balance", analytics.KPI.WarehouseBalance},
	}
	excelreport.WriteRows(f, sheet, 5, kpiRows)
	excelreport.ApplyTable(f, sheet, 5, len(kpiRows), 2, styles, false)
	_ = f.SetCellStyle(sheet, "B6", "B7", styles.Money)
	_ = f.SetCellStyle(sheet, "B8", "B8", styles.Integer)
	_ = f.SetCellStyle(sheet, "B9", "B9", styles.Money)

	fleetRows := [][]any{
		{"Fleet metric", "Count"},
		{"Total", analytics.FleetStatus.Total},
		{"On trip", analytics.FleetStatus.OnTrip},
		{"In garage", analytics.FleetStatus.InGarage},
		{"In repair", analytics.FleetStatus.InRepair},
		{"In reserve", analytics.FleetStatus.InReserve},
	}
	excelreport.WriteRows(f, sheet, 12, fleetRows)
	excelreport.ApplyTable(f, sheet, 12, len(fleetRows), 2, styles, false)
	_ = f.SetCellStyle(sheet, "B13", "B17", styles.Integer)

	repairRows := [][]any{
		{"Repair type", "Percent"},
		{"Planned TO", analytics.RepairBreakdown.PlannedTOPercent},
		{"Unplanned repair", analytics.RepairBreakdown.UnplannedPercent},
		{"Accident", analytics.RepairBreakdown.AccidentPercent},
	}
	excelreport.WriteRows(f, sheet, 20, repairRows)
	excelreport.ApplyTable(f, sheet, 20, len(repairRows), 2, styles, false)
	_ = f.SetCellStyle(sheet, "B21", "B23", styles.Integer)

	excelreport.SetWidths(f, sheet, []float64{28, 22})
}

func writeAnalyticsMonthlySheet(f *excelize.File, analytics *dto.AnalyticsDashboardResponse, styles excelreport.Styles) {
	const sheet = "Monthly expenses"
	_, _ = f.NewSheet(sheet)
	rows := [][]any{{"Month", "Fuel", "Repairs", "Parts", "Total"}}
	for _, item := range analytics.MonthlyExpenses {
		rows = append(rows, []any{item.Month, item.Fuel, item.Repairs, item.Parts, item.Fuel + item.Repairs + item.Parts})
	}
	excelreport.WriteRows(f, sheet, 1, rows)
	excelreport.ApplyTable(f, sheet, 1, len(rows), 5, styles, true)
	excelreport.FreezeBelow(f, sheet, 1)
	excelreport.SetWidths(f, sheet, []float64{16, 18, 18, 18, 18})
	if len(rows) > 1 {
		_ = f.SetCellStyle(sheet, "B2", fmt.Sprintf("E%d", len(rows)), styles.Money)
	}
}

func writeAnalyticsCriticalPartsSheet(f *excelize.File, analytics *dto.AnalyticsDashboardResponse, styles excelreport.Styles) {
	const sheet = "Critical parts"
	_, _ = f.NewSheet(sheet)
	rows := [][]any{{"Name", "Remaining percent", "Status"}}
	for _, item := range analytics.CriticalParts {
		rows = append(rows, []any{item.Name, item.RemainingPercent, item.Status})
	}
	excelreport.WriteRows(f, sheet, 1, rows)
	excelreport.ApplyTable(f, sheet, 1, len(rows), 3, styles, true)
	excelreport.FreezeBelow(f, sheet, 1)
	excelreport.SetWidths(f, sheet, []float64{44, 20, 16})
	if len(rows) > 1 {
		_ = f.SetCellStyle(sheet, "B2", fmt.Sprintf("B%d", len(rows)), styles.Integer)
	}
}

func setupDashboardPDFFont(pdf *gofpdf.Fpdf) string {
	// Put any licensed UTF-8 TTF file into one of these paths if you need perfect Cyrillic in PDFs.
	// The code works without it, but built-in PDF fonts are limited.
	candidates := []string{
		"assets/fonts/DejaVuSans.ttf",
		"storage/fonts/DejaVuSans.ttf",
		"fonts/DejaVuSans.ttf",
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			pdf.AddUTF8Font("Unicode", "", path)
			pdf.AddUTF8Font("Unicode", "B", path)
			return "Unicode"
		}
	}
	return "Helvetica"
}

func writePDFSectionTitle(pdf *gofpdf.Fpdf, fontName string, title string) {
	pdf.Ln(3)
	pdf.SetFont(fontName, "B", 12)
	pdf.CellFormat(0, 7, title, "", 1, "L", false, 0, "")
	pdf.SetFont(fontName, "", 9)
}

func writePDFSimpleRows(pdf *gofpdf.Fpdf, fontName string, rows [][]string) {
	pdf.SetFont(fontName, "", 9)
	for _, row := range rows {
		pdf.CellFormat(70, 6, row[0], "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 6, row[1], "1", 1, "R", false, 0, "")
	}
}

func writePDFTable(pdf *gofpdf.Fpdf, fontName string, widths []float64, headers []string, rows [][]string) {
	pdf.SetFont(fontName, "B", 8)
	for i, header := range headers {
		pdf.CellFormat(widths[i], 7, header, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont(fontName, "", 7)
	for _, row := range rows {
		for i, value := range row {
			align := "L"
			if i == 0 || i >= len(row)-4 {
				align = "R"
			}
			pdf.CellFormat(widths[i], 6, truncatePDFCell(value, 28), "1", 0, align, false, 0, "")
		}
		pdf.Ln(-1)
	}
}

func monthlyRows(stats *dto.DashboardStatsResponse) [][]string {
	rows := make([][]string, 0, len(stats.MonthlyExpenses))
	for _, item := range stats.MonthlyExpenses {
		rows = append(rows, []string{item.Month, money(item.FuelExpenses), money(item.RepairExpenses), money(item.TotalExpenses)})
	}
	return rows
}

func fleetRows(stats *dto.DashboardStatsResponse) [][]string {
	rows := make([][]string, 0, len(stats.FleetStatus.Items)+1)
	for _, item := range stats.FleetStatus.Items {
		rows = append(rows, []string{strconv.FormatInt(item.StatusID, 10), item.StatusName, strconv.FormatInt(item.Count, 10)})
	}
	rows = append(rows, []string{"", "Total", strconv.FormatInt(stats.FleetStatus.Total, 10)})
	return rows
}

func stockRows(stats *dto.DashboardStatsResponse) [][]string {
	rows := make([][]string, 0, len(stats.WarehouseStock.Items))
	for _, item := range stats.WarehouseStock.Items {
		rows = append(rows, []string{
			strconv.FormatInt(item.ID, 10),
			item.PartID,
			item.Name,
			strconv.FormatInt(item.Quantity, 10),
			money(item.Price),
			money(item.TotalValue),
			strconv.FormatBool(item.IsConsumable),
		})
	}
	return rows
}

func dashboardExportFilename(ext string) string {
	return fmt.Sprintf("dashboard_stats_%s.%s", time.Now().Format("20060102_150405"), ext)
}

func analyticsExportFilename(period string, ext string) string {
	return fmt.Sprintf("analytics_%s_%s.%s", period, time.Now().Format("20060102_150405"), ext)
}

func money(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func percent(v float64) string {
	return fmt.Sprintf("%.2f%%", v)
}

func truncatePDFCell(value string, maxRunes int) string {
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	if maxRunes <= 1 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-1]) + "…"
}
