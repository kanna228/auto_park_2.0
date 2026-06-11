package auto_park

import (
	"context"
	"database/sql"
	"net/http"
	"os"

	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/audit_log_module"
	auditlogrepository "auto_park/modules/audit_log_module/repository"
	auditlogservice "auto_park/modules/audit_log_module/service"
	"auto_park/modules/dashboard_module"
	"auto_park/modules/fuel_module"
	"auto_park/modules/fuel_norm_module"
	"auto_park/modules/incident_module"
	"auto_park/modules/invoice_module"
	invoicerepository "auto_park/modules/invoice_module/repository"
	invoiceservice "auto_park/modules/invoice_module/service"
	"auto_park/modules/maintenance_execution_module"
	"auto_park/modules/maintenance_schedule_module"
	"auto_park/modules/notification_module"
	notificationjobs "auto_park/modules/notification_module/jobs"
	notificationrepository "auto_park/modules/notification_module/repository"
	notificationservice "auto_park/modules/notification_module/service"
	notificationws "auto_park/modules/notification_module/websocket"
	"auto_park/modules/tripsheet_module"
	"auto_park/modules/user_module"
	"auto_park/modules/vehicle_module"
	"auto_park/modules/warehouse_module"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(cfg *config.Config, db *sql.DB) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Важно: CORS должен быть ДО всех модулей и ДО JWT middleware.
	// Иначе preflight OPTIONS запросы от фронта будут падать на авторизации.
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.PaginationAliasMiddleware())
	r.Use(middleware.DeleteForeignKeyConflictMiddleware())

	_ = os.MkdirAll(cfg.Uploads.RootDir, 0o755)

	r.Static("/static", cfg.Uploads.RootDir)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/healthz", func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"ok": false,
				"db": "not ok",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ok":      true,
			"service": cfg.Service.Name,
			"port":    cfg.Service.Port,
		})
	})

	auditLogRepo := auditlogrepository.NewAuditLogRepository(db)
	auditLogSvc := auditlogservice.NewAuditLogService(auditLogRepo)

	user_module.RegisterRoutes(r, cfg, db, auditLogSvc)
	vehicle_module.RegisterRoutes(r, cfg, db, auditLogSvc)
	tripsheet_module.RegisterRoutes(r, cfg, db, auditLogSvc)

	notificationHub := notificationws.NewHub()
	notificationRepo := notificationrepository.NewNotificationRepository(db)
	notificationSvc := notificationservice.NewNotificationService(notificationRepo, notificationHub)
	incident_module.RegisterRoutes(r, cfg, db, notificationSvc, auditLogSvc)
	notification_module.RegisterRoutes(r, cfg, notificationSvc, notificationHub)

	replacementReminderJob := notificationjobs.NewVehiclePartReplacementReminderJob(db, notificationSvc)
	replacementReminderJob.Start(context.Background())

	dashboard_module.RegisterRoutes(r, cfg, db)
	fuel_module.RegisterRoutes(r, cfg, db)
	fuel_norm_module.RegisterRoutes(r, cfg, db)
	maintenance_schedule_module.RegisterRoutes(r, cfg, db)
	maintenance_execution_module.RegisterRoutes(r, cfg, db)

	audit_log_module.RegisterRoutes(r, cfg, auditLogSvc)

	invoiceRepo := invoicerepository.NewInvoiceRepository(db)
	invoiceSvc := invoiceservice.NewInvoiceService(invoiceRepo)
	invoice_module.RegisterRoutes(r, cfg, invoiceSvc)

	warehouse_module.RegisterRoutes(r, cfg, db, notificationSvc, auditLogSvc, invoiceSvc)

	return r
}
