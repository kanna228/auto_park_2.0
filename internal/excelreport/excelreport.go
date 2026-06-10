package excelreport

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

type Styles struct {
	Title    int
	Header   int
	Body     int
	Money    int
	Integer  int
	Date     int
	DateTime int
}

func NewStyles(f *excelize.File) Styles {
	border := []excelize.Border{
		{Type: "left", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1},
		{Type: "top", Color: "000000", Style: 1},
		{Type: "bottom", Color: "000000", Style: 1},
	}
	title, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "1B2559"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	header, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1B2559"}, Pattern: 1},
		Border:    border,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	body, _ := f.NewStyle(&excelize.Style{
		Border:    border,
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})
	money, _ := f.NewStyle(&excelize.Style{
		NumFmt:    3,
		Border:    border,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	integer, _ := f.NewStyle(&excelize.Style{
		NumFmt:    3,
		Border:    border,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	date, _ := f.NewStyle(&excelize.Style{
		NumFmt:    14,
		Border:    border,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	dateTime, _ := f.NewStyle(&excelize.Style{
		NumFmt:    22,
		Border:    border,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	return Styles{Title: title, Header: header, Body: body, Money: money, Integer: integer, Date: date, DateTime: dateTime}
}

func WriteRows(f *excelize.File, sheet string, startRow int, rows [][]any) {
	for rowIdx, row := range rows {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, startRow+rowIdx)
			_ = f.SetCellValue(sheet, cell, value)
		}
	}
}

func ApplyTable(f *excelize.File, sheet string, startRow int, rowCount int, colCount int, styles Styles, withFilter bool) {
	if rowCount <= 0 || colCount <= 0 {
		return
	}

	endRow := startRow + rowCount - 1
	headerStart, _ := excelize.CoordinatesToCellName(1, startRow)
	headerEnd, _ := excelize.CoordinatesToCellName(colCount, startRow)
	_ = f.SetCellStyle(sheet, headerStart, headerEnd, styles.Header)
	_ = f.SetRowHeight(sheet, startRow, 24)

	if endRow > startRow {
		bodyStart, _ := excelize.CoordinatesToCellName(1, startRow+1)
		bodyEnd, _ := excelize.CoordinatesToCellName(colCount, endRow)
		_ = f.SetCellStyle(sheet, bodyStart, bodyEnd, styles.Body)
	}

	if withFilter && endRow > startRow {
		ref := fmt.Sprintf("%s:%s", headerStart, mustCell(colCount, endRow))
		_ = f.AutoFilter(sheet, ref, []excelize.AutoFilterOptions{})
	}
}

func FreezeBelow(f *excelize.File, sheet string, row int) {
	topLeft, _ := excelize.CoordinatesToCellName(1, row+1)
	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      row,
		TopLeftCell: topLeft,
		ActivePane:  "bottomLeft",
	})
}

func SetWidths(f *excelize.File, sheet string, widths []float64) {
	for i, width := range widths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetColWidth(sheet, col, col, width)
	}
}

func mustCell(col int, row int) string {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	return cell
}
