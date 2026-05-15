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

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
