package dto

type DashboardStatsResponse struct {
	Week            DashboardWeekRange            `json:"week"`
	Cards           DashboardCards                `json:"cards"`
	MonthlyExpenses []DashboardMonthlyExpenseItem `json:"monthly_expenses"`
	FleetStatus     DashboardFleetStatusBlock     `json:"fleet_status"`
	WarehouseStock  DashboardWarehouseStockBlock  `json:"warehouse_stock"`
}

type DashboardWeekRange struct {
	DateFrom string `json:"date_from" example:"2026-05-11"`
	DateTo   string `json:"date_to" example:"2026-05-17"`
}

type DashboardMoneyMetric struct {
	Current       float64 `json:"current" example:"620000"`
	Previous      float64 `json:"previous" example:"590000"`
	ChangePercent float64 `json:"change_percent" example:"5.08"`
	Trend         string  `json:"trend" example:"up"`
}

type DashboardCountMetric struct {
	Current       int64   `json:"current" example:"47"`
	Previous      int64   `json:"previous" example:"42"`
	ChangePercent float64 `json:"change_percent" example:"11.9"`
	Trend         string  `json:"trend" example:"up"`
}

type DashboardWarehouseBalanceMetric struct {
	Current       float64 `json:"current" example:"3100000"`
	CriticalCount int64   `json:"critical_count" example:"2"`
}

type DashboardCards struct {
	FuelExpenses   DashboardMoneyMetric            `json:"fuel_expenses"`
	RepairExpenses DashboardMoneyMetric            `json:"repair_expenses"`
	IssuedParts    DashboardIssuedPartsMetric      `json:"issued_parts"`
	WarehouseStock DashboardWarehouseBalanceMetric `json:"warehouse_stock"`
}

type DashboardIssuedPartsMetric struct {
	DashboardCountMetric
	PendingRequests int64 `json:"pending_requests" example:"3"`
}

type DashboardMonthlyExpenseItem struct {
	Month          string  `json:"month" example:"2026-05"`
	FuelExpenses   float64 `json:"fuel_expenses" example:"620000"`
	RepairExpenses float64 `json:"repair_expenses" example:"280000"`
	TotalExpenses  float64 `json:"total_expenses" example:"900000"`
}

type DashboardFleetStatusBlock struct {
	Total int64                      `json:"total" example:"48"`
	Items []DashboardFleetStatusItem `json:"items"`
}

type DashboardFleetStatusItem struct {
	StatusID   int64  `json:"status_id" example:"1"`
	StatusName string `json:"status_name" example:"На выезде"`
	Count      int64  `json:"count" example:"26"`
}

type DashboardWarehouseStockBlock struct {
	Limit         int                           `json:"limit" example:"7"`
	TotalValue    float64                       `json:"total_value" example:"3100000"`
	CriticalCount int64                         `json:"critical_count" example:"2"`
	Items         []DashboardWarehouseStockItem `json:"items"`
}

type DashboardWarehouseStockItem struct {
	ID           int64   `json:"id" example:"1"`
	PartID       string  `json:"part_id" example:"BRK-PAD-001"`
	Name         string  `json:"name" example:"Brake Pad Front"`
	Category     string  `json:"category" example:"brake_system"`
	Quantity     int64   `json:"quantity" example:"50"`
	Price        float64 `json:"price" example:"25000"`
	TotalValue   float64 `json:"total_value" example:"1250000"`
	IsConsumable bool    `json:"is_consumable" example:"false"`
}

type AnalyticsDashboardResponse struct {
	Period          string                    `json:"period"`
	KPI             AnalyticsKPI              `json:"kpi"`
	MonthlyExpenses []AnalyticsMonthlyExpense `json:"monthly_expenses"`
	FleetStatus     AnalyticsFleetStatus      `json:"fleet_status"`
	RepairBreakdown AnalyticsRepairBreakdown  `json:"repair_breakdown"`
	CriticalParts   []AnalyticsCriticalPart   `json:"critical_parts"`
}

type AnalyticsKPI struct {
	FuelCost         float64 `json:"fuel_cost"`
	RepairCost       float64 `json:"repair_cost"`
	PartsIssued      int64   `json:"parts_issued"`
	WarehouseBalance float64 `json:"warehouse_balance"`
}

type AnalyticsMonthlyExpense struct {
	Month   string  `json:"month"`
	Fuel    float64 `json:"fuel"`
	Repairs float64 `json:"repairs"`
	Parts   float64 `json:"parts"`
}

type AnalyticsFleetStatus struct {
	Total     int64 `json:"total"`
	OnTrip    int64 `json:"on_trip"`
	InGarage  int64 `json:"in_garage"`
	InRepair  int64 `json:"in_repair"`
	InReserve int64 `json:"in_reserve"`
}

type AnalyticsRepairBreakdown struct {
	PlannedTOPercent int64 `json:"planned_to_percent"`
	UnplannedPercent int64 `json:"unplanned_percent"`
	AccidentPercent  int64 `json:"accident_percent"`
}

type AnalyticsCriticalPart struct {
	Name             string `json:"name"`
	RemainingPercent int64  `json:"remaining_percent"`
	Status           string `json:"status"`
}
