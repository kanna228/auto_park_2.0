package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"auto_park/modules/dashboard_module/dto"
)

type DashboardRepository interface {
	GetStats(ctx context.Context) (*dto.DashboardStatsResponse, error)
	GetExportStats(ctx context.Context) (*dto.DashboardStatsResponse, error)
	GetAnalytics(ctx context.Context, period string) (*dto.AnalyticsDashboardResponse, error)
}

type dashboardRepo struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) DashboardRepository {
	return &dashboardRepo{db: db}
}

func (r *dashboardRepo) GetStats(ctx context.Context) (*dto.DashboardStatsResponse, error) {
	return r.getStats(ctx, 7)
}

func (r *dashboardRepo) GetExportStats(ctx context.Context) (*dto.DashboardStatsResponse, error) {
	// Stock limit = 0 means export ALL stock positions without dashboard card limit.
	return r.getStats(ctx, 0)
}

func (r *dashboardRepo) GetAnalytics(ctx context.Context, period string) (*dto.AnalyticsDashboardResponse, error) {
	period = normalizeAnalyticsPeriod(period)

	kpi, err := r.getAnalyticsKPI(ctx, period)
	if err != nil {
		return nil, err
	}
	monthly, err := r.getAnalyticsMonthlyExpenses(ctx)
	if err != nil {
		return nil, err
	}
	fleet, err := r.getAnalyticsFleetStatus(ctx)
	if err != nil {
		return nil, err
	}
	breakdown, err := r.getAnalyticsRepairBreakdown(ctx, period)
	if err != nil {
		return nil, err
	}
	criticalParts, err := r.getAnalyticsCriticalParts(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.AnalyticsDashboardResponse{
		Period:          period,
		KPI:             *kpi,
		MonthlyExpenses: monthly,
		FleetStatus:     *fleet,
		RepairBreakdown: *breakdown,
		CriticalParts:   criticalParts,
	}, nil
}

func (r *dashboardRepo) getStats(ctx context.Context, stockLimit int) (*dto.DashboardStatsResponse, error) {
	week, err := r.getWeekRange(ctx)
	if err != nil {
		return nil, err
	}

	fuel, err := r.getWeeklyMoneyMetric(ctx, "fuel_refills", "date", "price")
	if err != nil {
		return nil, err
	}

	repairs, err := r.getWeeklyMoneyMetric(ctx, "vehicle_part_installations", "installed_at", "total_price")
	if err != nil {
		return nil, err
	}

	issuedParts, err := r.getIssuedPartsMetric(ctx)
	if err != nil {
		return nil, err
	}

	warehouseBalance, err := r.getWarehouseBalanceMetric(ctx)
	if err != nil {
		return nil, err
	}

	monthly, err := r.getMonthlyExpenses(ctx)
	if err != nil {
		return nil, err
	}

	fleet, err := r.getFleetStatus(ctx)
	if err != nil {
		return nil, err
	}

	stock, err := r.getWarehouseStock(ctx, stockLimit)
	if err != nil {
		return nil, err
	}
	stock.TotalValue = warehouseBalance.Current
	stock.CriticalCount = warehouseBalance.CriticalCount

	return &dto.DashboardStatsResponse{
		Week: *week,
		Cards: dto.DashboardCards{
			FuelExpenses:   *fuel,
			RepairExpenses: *repairs,
			IssuedParts:    *issuedParts,
			WarehouseStock: *warehouseBalance,
		},
		MonthlyExpenses: monthly,
		FleetStatus:     *fleet,
		WarehouseStock:  *stock,
	}, nil
}

func (r *dashboardRepo) getAnalyticsKPI(ctx context.Context, period string) (*dto.AnalyticsKPI, error) {
	const q = `
		WITH bounds AS (
			SELECT
				date_trunc($1, CURRENT_DATE::timestamp) AS start_at,
				CASE
					WHEN $1 = 'day' THEN INTERVAL '1 day'
					WHEN $1 = 'week' THEN INTERVAL '7 days'
					WHEN $1 = 'month' THEN INTERVAL '1 month'
					WHEN $1 = 'quarter' THEN INTERVAL '3 months'
					WHEN $1 = 'year' THEN INTERVAL '1 year'
					ELSE INTERVAL '7 days'
				END AS span
		)
		SELECT
			COALESCE((SELECT SUM(price)::float8 FROM fuel_refills, bounds WHERE date >= bounds.start_at AND date < bounds.start_at + bounds.span), 0),
			COALESCE((SELECT SUM(total_price)::float8 FROM vehicle_part_installations, bounds WHERE installed_at >= bounds.start_at AND installed_at < bounds.start_at + bounds.span), 0),
			COALESCE((SELECT SUM(quantity)::bigint FROM vehicle_part_installations, bounds WHERE installed_at >= bounds.start_at AND installed_at < bounds.start_at + bounds.span), 0),
			COALESCE((SELECT SUM(quantity * price)::float8 FROM parts_catalog), 0);
	`

	var item dto.AnalyticsKPI
	if err := r.db.QueryRowContext(ctx, q, period).Scan(&item.FuelCost, &item.RepairCost, &item.PartsIssued, &item.WarehouseBalance); err != nil {
		return nil, fmt.Errorf("analytics kpi: %w", err)
	}
	item.FuelCost = round2(item.FuelCost)
	item.RepairCost = round2(item.RepairCost)
	item.WarehouseBalance = round2(item.WarehouseBalance)
	return &item, nil
}

func (r *dashboardRepo) getAnalyticsMonthlyExpenses(ctx context.Context) ([]dto.AnalyticsMonthlyExpense, error) {
	const q = `
		WITH months AS (
			SELECT (date_trunc('month', CURRENT_DATE)::date - (n || ' months')::interval)::date AS month_start
			FROM generate_series(5, 0, -1) AS gs(n)
		), fuel AS (
			SELECT date_trunc('month', date)::date AS month_start, SUM(price)::float8 AS value
			FROM fuel_refills
			WHERE date >= (SELECT MIN(month_start) FROM months)
			GROUP BY 1
		), repairs AS (
			SELECT date_trunc('month', installed_at)::date AS month_start, SUM(total_price)::float8 AS value
			FROM vehicle_part_installations
			WHERE installed_at >= (SELECT MIN(month_start) FROM months)
			GROUP BY 1
		), parts AS (
			SELECT date_trunc('month', m.created_at)::date AS month_start, SUM(m.quantity * p.price)::float8 AS value
			FROM part_stock_movements m
			INNER JOIN parts_catalog p ON p.id = m.part_id
			WHERE m.type = 'issue'
			  AND m.created_at >= (SELECT MIN(month_start) FROM months)
			GROUP BY 1
		)
		SELECT
			to_char(m.month_start, 'YYYY-MM') AS month,
			COALESCE(f.value, 0)::float8 AS fuel,
			COALESCE(r.value, 0)::float8 AS repairs,
			COALESCE(p.value, 0)::float8 AS parts
		FROM months m
		LEFT JOIN fuel f ON f.month_start = m.month_start
		LEFT JOIN repairs r ON r.month_start = m.month_start
		LEFT JOIN parts p ON p.month_start = m.month_start
		ORDER BY m.month_start ASC;
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("analytics monthly expenses: %w", err)
	}
	defer rows.Close()

	items := make([]dto.AnalyticsMonthlyExpense, 0, 6)
	for rows.Next() {
		var item dto.AnalyticsMonthlyExpense
		if err := rows.Scan(&item.Month, &item.Fuel, &item.Repairs, &item.Parts); err != nil {
			return nil, fmt.Errorf("analytics monthly expenses scan: %w", err)
		}
		item.Fuel = round2(item.Fuel)
		item.Repairs = round2(item.Repairs)
		item.Parts = round2(item.Parts)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("analytics monthly expenses rows: %w", err)
	}
	return items, nil
}

func (r *dashboardRepo) getAnalyticsFleetStatus(ctx context.Context) (*dto.AnalyticsFleetStatus, error) {
	const q = `
		WITH open_vehicles AS (
			SELECT DISTINCT vehicle_id
			FROM tripsheets
			WHERE vehicle_id IS NOT NULL
			  AND end_time IS NULL
			  AND status_id NOT IN (4, 5)
		)
		SELECT
			COUNT(v.id)::bigint AS total,
			COUNT(v.id) FILTER (WHERE ov.vehicle_id IS NOT NULL)::bigint AS on_trip,
			COUNT(v.id) FILTER (WHERE v.status_id = 1 AND ov.vehicle_id IS NULL)::bigint AS in_garage,
			COUNT(v.id) FILTER (WHERE v.status_id IN (2, 4))::bigint AS in_repair,
			COUNT(v.id) FILTER (WHERE v.status_id IN (3, 5))::bigint AS in_reserve
		FROM vehicles v
		LEFT JOIN open_vehicles ov ON ov.vehicle_id = v.id;
	`

	var item dto.AnalyticsFleetStatus
	if err := r.db.QueryRowContext(ctx, q).Scan(&item.Total, &item.OnTrip, &item.InGarage, &item.InRepair, &item.InReserve); err != nil {
		return nil, fmt.Errorf("analytics fleet status: %w", err)
	}
	return &item, nil
}

func (r *dashboardRepo) getAnalyticsRepairBreakdown(ctx context.Context, period string) (*dto.AnalyticsRepairBreakdown, error) {
	const q = `
		WITH bounds AS (
			SELECT
				date_trunc($1, CURRENT_DATE::timestamp) AS start_at,
				CASE
					WHEN $1 = 'day' THEN INTERVAL '1 day'
					WHEN $1 = 'week' THEN INTERVAL '7 days'
					WHEN $1 = 'month' THEN INTERVAL '1 month'
					WHEN $1 = 'quarter' THEN INTERVAL '3 months'
					WHEN $1 = 'year' THEN INTERVAL '1 year'
					ELSE INTERVAL '7 days'
				END AS span
		), counts AS (
			SELECT
				(SELECT COUNT(*)::float8 FROM maintenance_executions, bounds WHERE created_at >= bounds.start_at AND created_at < bounds.start_at + bounds.span) AS planned,
				(SELECT COUNT(*)::float8 FROM incidents, bounds WHERE incident_type_id <> 1 AND incident_date >= bounds.start_at::date AND incident_date < (bounds.start_at + bounds.span)::date) AS unplanned,
				(SELECT COUNT(*)::float8 FROM incidents, bounds WHERE incident_type_id = 1 AND incident_date >= bounds.start_at::date AND incident_date < (bounds.start_at + bounds.span)::date) AS accident
		)
		SELECT planned, unplanned, accident
		FROM counts;
	`

	var planned, unplanned, accident float64
	if err := r.db.QueryRowContext(ctx, q, period).Scan(&planned, &unplanned, &accident); err != nil {
		return nil, fmt.Errorf("analytics repair breakdown: %w", err)
	}
	total := planned + unplanned + accident
	if total <= 0 {
		return &dto.AnalyticsRepairBreakdown{}, nil
	}

	plannedPercent := int64(math.Round(planned / total * 100))
	unplannedPercent := int64(math.Round(unplanned / total * 100))
	accidentPercent := int64(100 - plannedPercent - unplannedPercent)
	if accidentPercent < 0 {
		accidentPercent = int64(math.Round(accident / total * 100))
	}

	return &dto.AnalyticsRepairBreakdown{
		PlannedTOPercent: plannedPercent,
		UnplannedPercent: unplannedPercent,
		AccidentPercent:  accidentPercent,
	}, nil
}

func (r *dashboardRepo) getAnalyticsCriticalParts(ctx context.Context) ([]dto.AnalyticsCriticalPart, error) {
	const q = `
		SELECT
			name,
			CASE
				WHEN min_stock_quantity <= 0 THEN CASE WHEN quantity > 0 THEN 100 ELSE 0 END
				ELSE LEAST(100, ROUND(quantity::numeric / NULLIF(min_stock_quantity * 2, 0)::numeric * 100))::bigint
			END AS remaining_percent,
			CASE
				WHEN quantity <= 0 OR (min_stock_quantity > 0 AND quantity <= min_stock_quantity) THEN 'critical'
				WHEN min_stock_quantity > 0 AND quantity <= min_stock_quantity * 2 THEN 'low'
				ELSE 'ok'
			END AS status
		FROM parts_catalog
		ORDER BY
			CASE
				WHEN quantity <= 0 OR (min_stock_quantity > 0 AND quantity <= min_stock_quantity) THEN 0
				WHEN min_stock_quantity > 0 AND quantity <= min_stock_quantity * 2 THEN 1
				ELSE 2
			END ASC,
			remaining_percent ASC,
			name ASC
		LIMIT 10;
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("analytics critical parts: %w", err)
	}
	defer rows.Close()

	items := make([]dto.AnalyticsCriticalPart, 0)
	for rows.Next() {
		var item dto.AnalyticsCriticalPart
		if err := rows.Scan(&item.Name, &item.RemainingPercent, &item.Status); err != nil {
			return nil, fmt.Errorf("analytics critical parts scan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("analytics critical parts rows: %w", err)
	}
	return items, nil
}

func (r *dashboardRepo) getWeekRange(ctx context.Context) (*dto.DashboardWeekRange, error) {
	const q = `
		SELECT
			date_trunc('week', CURRENT_DATE)::date,
			(date_trunc('week', CURRENT_DATE)::date + INTERVAL '6 days')::date;
	`

	var from time.Time
	var to time.Time
	if err := r.db.QueryRowContext(ctx, q).Scan(&from, &to); err != nil {
		return nil, fmt.Errorf("dashboard week range: %w", err)
	}

	return &dto.DashboardWeekRange{
		DateFrom: from.Format("2006-01-02"),
		DateTo:   to.Format("2006-01-02"),
	}, nil
}

func (r *dashboardRepo) getWeeklyMoneyMetric(ctx context.Context, table string, dateColumn string, amountColumn string) (*dto.DashboardMoneyMetric, error) {
	q := fmt.Sprintf(`
		WITH bounds AS (
			SELECT
				date_trunc('week', CURRENT_DATE)::date AS current_start,
				(date_trunc('week', CURRENT_DATE)::date - INTERVAL '7 days')::date AS previous_start
		)
		SELECT
			COALESCE(SUM(CASE WHEN %s >= current_start AND %s < current_start + INTERVAL '7 days' THEN %s ELSE 0 END), 0)::float8 AS current_value,
			COALESCE(SUM(CASE WHEN %s >= previous_start AND %s < current_start THEN %s ELSE 0 END), 0)::float8 AS previous_value
		FROM %s, bounds;
	`, dateColumn, dateColumn, amountColumn, dateColumn, dateColumn, amountColumn, table)

	var current float64
	var previous float64
	if err := r.db.QueryRowContext(ctx, q).Scan(&current, &previous); err != nil {
		return nil, fmt.Errorf("dashboard weekly money metric %s: %w", table, err)
	}

	return &dto.DashboardMoneyMetric{
		Current:       round2(current),
		Previous:      round2(previous),
		ChangePercent: percentChange(current, previous),
		Trend:         trend(current, previous),
	}, nil
}

func (r *dashboardRepo) getIssuedPartsMetric(ctx context.Context) (*dto.DashboardIssuedPartsMetric, error) {
	const q = `
		WITH bounds AS (
			SELECT
				date_trunc('week', CURRENT_DATE)::date AS current_start,
				(date_trunc('week', CURRENT_DATE)::date - INTERVAL '7 days')::date AS previous_start
		), issued AS (
			SELECT
				COALESCE(SUM(CASE WHEN installed_at >= current_start AND installed_at < current_start + INTERVAL '7 days' THEN quantity ELSE 0 END), 0)::bigint AS current_value,
				COALESCE(SUM(CASE WHEN installed_at >= previous_start AND installed_at < current_start THEN quantity ELSE 0 END), 0)::bigint AS previous_value
			FROM vehicle_part_installations, bounds
		), pending AS (
			SELECT COUNT(*)::bigint AS pending_count
			FROM part_requests pr
			INNER JOIN part_request_statuses s ON s.id = pr.status_id
			WHERE pr.is_deleted = FALSE
			  AND s.code = 'new'
		)
		SELECT issued.current_value, issued.previous_value, pending.pending_count
		FROM issued, pending;
	`

	var current int64
	var previous int64
	var pending int64
	if err := r.db.QueryRowContext(ctx, q).Scan(&current, &previous, &pending); err != nil {
		return nil, fmt.Errorf("dashboard issued parts metric: %w", err)
	}

	return &dto.DashboardIssuedPartsMetric{
		DashboardCountMetric: dto.DashboardCountMetric{
			Current:       current,
			Previous:      previous,
			ChangePercent: percentChange(float64(current), float64(previous)),
			Trend:         trend(float64(current), float64(previous)),
		},
		PendingRequests: pending,
	}, nil
}

func (r *dashboardRepo) getWarehouseBalanceMetric(ctx context.Context) (*dto.DashboardWarehouseBalanceMetric, error) {
	const q = `
		SELECT
			COALESCE(SUM(quantity * price), 0)::float8 AS total_value,
			COUNT(*) FILTER (WHERE quantity <= 5)::bigint AS critical_count
		FROM parts_catalog;
	`

	var total float64
	var critical int64
	if err := r.db.QueryRowContext(ctx, q).Scan(&total, &critical); err != nil {
		return nil, fmt.Errorf("dashboard warehouse balance metric: %w", err)
	}

	return &dto.DashboardWarehouseBalanceMetric{
		Current:       round2(total),
		CriticalCount: critical,
	}, nil
}

func (r *dashboardRepo) getMonthlyExpenses(ctx context.Context) ([]dto.DashboardMonthlyExpenseItem, error) {
	const q = `
		WITH months AS (
			SELECT (date_trunc('month', CURRENT_DATE)::date - (n || ' months')::interval)::date AS month_start
			FROM generate_series(5, 0, -1) AS gs(n)
		), fuel AS (
			SELECT date_trunc('month', date)::date AS month_start, SUM(price)::float8 AS value
			FROM fuel_refills
			WHERE date >= (SELECT MIN(month_start) FROM months)
			GROUP BY 1
		), repairs AS (
			SELECT date_trunc('month', installed_at)::date AS month_start, SUM(total_price)::float8 AS value
			FROM vehicle_part_installations
			WHERE installed_at >= (SELECT MIN(month_start) FROM months)
			GROUP BY 1
		)
		SELECT
			to_char(m.month_start, 'YYYY-MM') AS month,
			COALESCE(f.value, 0)::float8 AS fuel_expenses,
			COALESCE(r.value, 0)::float8 AS repair_expenses
		FROM months m
		LEFT JOIN fuel f ON f.month_start = m.month_start
		LEFT JOIN repairs r ON r.month_start = m.month_start
		ORDER BY m.month_start ASC;
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("dashboard monthly expenses: %w", err)
	}
	defer rows.Close()

	items := make([]dto.DashboardMonthlyExpenseItem, 0, 6)
	for rows.Next() {
		var item dto.DashboardMonthlyExpenseItem
		if err := rows.Scan(&item.Month, &item.FuelExpenses, &item.RepairExpenses); err != nil {
			return nil, fmt.Errorf("dashboard monthly expenses scan: %w", err)
		}
		item.FuelExpenses = round2(item.FuelExpenses)
		item.RepairExpenses = round2(item.RepairExpenses)
		item.TotalExpenses = round2(item.FuelExpenses + item.RepairExpenses)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard monthly expenses rows: %w", err)
	}
	return items, nil
}

func (r *dashboardRepo) getFleetStatus(ctx context.Context) (*dto.DashboardFleetStatusBlock, error) {
	const q = `
		SELECT
			vs.id,
			vs.name,
			COUNT(v.id)::bigint AS cnt
		FROM vehicle_status vs
		LEFT JOIN vehicles v ON v.status_id = vs.id
		GROUP BY vs.id, vs.name
		ORDER BY vs.id ASC;
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("dashboard fleet status: %w", err)
	}
	defer rows.Close()

	items := make([]dto.DashboardFleetStatusItem, 0)
	var total int64
	for rows.Next() {
		var item dto.DashboardFleetStatusItem
		if err := rows.Scan(&item.StatusID, &item.StatusName, &item.Count); err != nil {
			return nil, fmt.Errorf("dashboard fleet status scan: %w", err)
		}
		total += item.Count
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard fleet status rows: %w", err)
	}

	return &dto.DashboardFleetStatusBlock{Total: total, Items: items}, nil
}

func (r *dashboardRepo) getWarehouseStock(ctx context.Context, limit int) (*dto.DashboardWarehouseStockBlock, error) {
	query := `
		SELECT
			id,
			part_id,
			name,
			category,
			quantity,
			price::float8,
			(quantity * price)::float8 AS total_value,
			is_consumable
		FROM parts_catalog
		ORDER BY quantity ASC, total_value DESC, id ASC
	`

	args := []any{}
	resultLimit := limit
	if limit > 0 {
		if limit > 50 {
			limit = 50
		}
		resultLimit = limit
		query += ` LIMIT $1`
		args = append(args, limit)
	} else {
		// 0 in response means no row limit was applied.
		resultLimit = 0
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("dashboard warehouse stock: %w", err)
	}
	defer rows.Close()

	items := make([]dto.DashboardWarehouseStockItem, 0)
	for rows.Next() {
		var item dto.DashboardWarehouseStockItem
		if err := rows.Scan(&item.ID, &item.PartID, &item.Name, &item.Category, &item.Quantity, &item.Price, &item.TotalValue, &item.IsConsumable); err != nil {
			return nil, fmt.Errorf("dashboard warehouse stock scan: %w", err)
		}
		item.Price = round2(item.Price)
		item.TotalValue = round2(item.TotalValue)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard warehouse stock rows: %w", err)
	}

	return &dto.DashboardWarehouseStockBlock{Limit: resultLimit, Items: items}, nil
}

func percentChange(current float64, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return round2(((current - previous) / previous) * 100)
}

func trend(current float64, previous float64) string {
	switch {
	case current > previous:
		return "up"
	case current < previous:
		return "down"
	default:
		return "same"
	}
}

func normalizeAnalyticsPeriod(period string) string {
	switch period {
	case "day", "week", "month", "quarter", "year":
		return period
	default:
		return "week"
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
