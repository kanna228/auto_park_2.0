package dto

type FuelReportFilter struct {
	DateFrom *string `form:"date_from" example:"2026-04-01"`
	DateTo   *string `form:"date_to" example:"2026-04-30"`
	Format   string  `form:"format" example:"pdf"`
}
