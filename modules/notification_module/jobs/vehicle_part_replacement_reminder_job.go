package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	notificationservice "auto_park/modules/notification_module/service"
)

const (
	defaultReplacementReminderInterval = 6 * time.Hour
)

type VehiclePartReplacementReminderJob struct {
	db       *sql.DB
	notify   notificationservice.NotificationService
	interval time.Duration
}

type dueVehiclePartReplacement struct {
	InstallationID       int64
	PartID               int64
	PartCatalogCode      string
	PartName             string
	PartCategory         string
	VehicleID            int64
	VehicleStateNumber   string
	VehicleBrandModel    string
	PlannedReplacementAt time.Time
	InstalledByUserID    int64
	InstallerEmail       *string
	InstallerFullName    *string
	DaysBefore           int
}

func NewVehiclePartReplacementReminderJob(db *sql.DB, notify notificationservice.NotificationService) *VehiclePartReplacementReminderJob {
	return &VehiclePartReplacementReminderJob{
		db:       db,
		notify:   notify,
		interval: defaultReplacementReminderInterval,
	}
}

func (j *VehiclePartReplacementReminderJob) Start(ctx context.Context) {
	if j == nil || j.db == nil || j.notify == nil {
		return
	}
	if j.interval <= 0 {
		j.interval = defaultReplacementReminderInterval
	}

	go func() {
		j.runSafely(ctx)

		ticker := time.NewTicker(j.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Printf("[replacement-reminder-job] stopped: %v", ctx.Err())
				return
			case <-ticker.C:
				j.runSafely(ctx)
			}
		}
	}()
}

func (j *VehiclePartReplacementReminderJob) runSafely(ctx context.Context) {
	if err := j.RunOnce(ctx); err != nil {
		log.Printf("[replacement-reminder-job] error: %v", err)
	}
}

func (j *VehiclePartReplacementReminderJob) RunOnce(ctx context.Context) error {
	items, err := j.findDueInstallations(ctx)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		log.Printf("[replacement-reminder-job] no due replacements")
		return nil
	}

	created := 0
	for _, item := range items {
		n, err := j.notifyForInstallation(ctx, item)
		if err != nil {
			log.Printf("[replacement-reminder-job] installation_id=%d notify error: %v", item.InstallationID, err)
			continue
		}
		created += n
	}

	log.Printf("[replacement-reminder-job] checked=%d created=%d", len(items), created)
	return nil
}

func (j *VehiclePartReplacementReminderJob) findDueInstallations(ctx context.Context) ([]dueVehiclePartReplacement, error) {
	const q = `
		SELECT
			vpi.id,
			vpi.part_id,
			p.part_id AS part_catalog_code,
			p.name AS part_name,
			p.category AS part_category,
			vpi.vehicle_id,
			v.state_number AS vehicle_state_number,
			v.brand_model AS vehicle_brand_model,
			vpi.planned_replacement_at,
			vpi.installed_by_user_id,
			u.email AS installer_email,
			NULLIF(TRIM(CONCAT_WS(' ', u.last_name, u.first_name, u.middle_name)), '') AS installer_full_name,
			CASE
				WHEN vpi.planned_replacement_at = CURRENT_DATE + 7 THEN 7
				WHEN vpi.planned_replacement_at = CURRENT_DATE THEN 0
				ELSE -1
			END AS days_before
		FROM vehicle_part_installations vpi
		INNER JOIN parts_catalog p ON p.id = vpi.part_id
		INNER JOIN vehicles v ON v.id = vpi.vehicle_id
		LEFT JOIN users u ON u.id = vpi.installed_by_user_id
		WHERE vpi.is_active = TRUE
		  AND vpi.planned_replacement_at IS NOT NULL
		  AND vpi.planned_replacement_at IN (CURRENT_DATE, CURRENT_DATE + 7)
		ORDER BY vpi.planned_replacement_at ASC, vpi.id ASC;
	`

	rows, err := j.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("find due vehicle part replacements: %w", err)
	}
	defer rows.Close()

	items := make([]dueVehiclePartReplacement, 0)
	for rows.Next() {
		var item dueVehiclePartReplacement
		var installerEmail sql.NullString
		var installerFullName sql.NullString

		if err := rows.Scan(
			&item.InstallationID,
			&item.PartID,
			&item.PartCatalogCode,
			&item.PartName,
			&item.PartCategory,
			&item.VehicleID,
			&item.VehicleStateNumber,
			&item.VehicleBrandModel,
			&item.PlannedReplacementAt,
			&item.InstalledByUserID,
			&installerEmail,
			&installerFullName,
			&item.DaysBefore,
		); err != nil {
			return nil, fmt.Errorf("find due vehicle part replacements scan: %w", err)
		}

		if installerEmail.Valid {
			item.InstallerEmail = &installerEmail.String
		}
		if installerFullName.Valid {
			item.InstallerFullName = &installerFullName.String
		}

		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("find due vehicle part replacements rows: %w", err)
	}

	return items, nil
}

func (j *VehiclePartReplacementReminderJob) notifyForInstallation(ctx context.Context, item dueVehiclePartReplacement) (int, error) {
	typeCode, title, message, reminderKind := buildReplacementReminderText(item)
	dedupKey := buildReplacementReminderDedupKey(item)
	contextData := buildReplacementReminderContext(item, reminderKind)

	created := 0
	managerItems, err := j.notify.CreateForRoleOnce(
		ctx,
		notificationservice.WarehouseManagerRoleCode,
		notificationservice.WarehouseManagerFallbackRoleID,
		dedupKey,
		typeCode,
		title,
		message,
		contextData,
	)
	if err != nil {
		return created, fmt.Errorf("notify warehouse managers: %w", err)
	}
	created += len(managerItems)

	mechanicItem, err := j.notify.CreateForUserOnce(
		ctx,
		item.InstalledByUserID,
		dedupKey,
		typeCode,
		title,
		message,
		contextData,
	)
	if err != nil {
		return created, fmt.Errorf("notify installer mechanic: %w", err)
	}
	if mechanicItem != nil {
		created++
	}

	return created, nil
}

func buildReplacementReminderText(item dueVehiclePartReplacement) (typeCode string, title string, message string, reminderKind string) {
	vehicleLabel := strings.TrimSpace(item.VehicleStateNumber)
	if item.VehicleBrandModel != "" {
		vehicleLabel = strings.TrimSpace(vehicleLabel + " " + item.VehicleBrandModel)
	}
	if vehicleLabel == "" {
		vehicleLabel = fmt.Sprintf("vehicle_id=%d", item.VehicleID)
	}

	dateStr := item.PlannedReplacementAt.Format("2006-01-02")
	partLabel := strings.TrimSpace(item.PartName)
	if partLabel == "" {
		partLabel = fmt.Sprintf("part_id=%d", item.PartID)
	}

	if item.DaysBefore == 7 {
		return notificationservice.NotificationTypeVehiclePartReplacement7Days,
			"Скоро плановая замена детали",
			fmt.Sprintf("Через 7 дней, %s, нужно заменить деталь %q на машине %s.", dateStr, partLabel, vehicleLabel),
			"7_days"
	}

	return notificationservice.NotificationTypeVehiclePartReplacementToday,
		"Сегодня плановая замена детали",
		fmt.Sprintf("Сегодня, %s, нужно заменить деталь %q на машине %s.", dateStr, partLabel, vehicleLabel),
		"today"
}

func buildReplacementReminderDedupKey(item dueVehiclePartReplacement) string {
	kind := "today"
	if item.DaysBefore == 7 {
		kind = "7_days"
	}
	return fmt.Sprintf(
		"vehicle_part_replacement:%s:installation:%d:planned:%s",
		kind,
		item.InstallationID,
		item.PlannedReplacementAt.Format("2006-01-02"),
	)
}

func buildReplacementReminderContext(item dueVehiclePartReplacement, reminderKind string) map[string]any {
	return map[string]any{
		"module":                       "warehouse",
		"entity":                       "vehicle_part_installation",
		"action":                       "planned_replacement_reminder",
		"reminder_kind":                reminderKind,
		"days_before":                  item.DaysBefore,
		"installation_id":              item.InstallationID,
		"vehicle_part_installation_id": item.InstallationID,
		"part_id":                      item.PartID,
		"part_catalog_code":            item.PartCatalogCode,
		"part_name":                    item.PartName,
		"part_category":                item.PartCategory,
		"vehicle_id":                   item.VehicleID,
		"vehicle_state_number":         item.VehicleStateNumber,
		"vehicle_brand_model":          item.VehicleBrandModel,
		"planned_replacement_at":       item.PlannedReplacementAt.Format("2006-01-02"),
		"installed_by_user_id":         item.InstalledByUserID,
		"installer_email":              item.InstallerEmail,
		"installer_full_name":          item.InstallerFullName,
		"frontend_hint": map[string]any{
			"screen": "vehicle_part_installation_details",
			"params": map[string]any{
				"id":         item.InstallationID,
				"vehicle_id": item.VehicleID,
				"part_id":    item.PartID,
			},
		},
	}
}
