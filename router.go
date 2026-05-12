package auto_park

import (
	"database/sql"
	"net/http"
	"os"

	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/fuel_module"
	"auto_park/modules/incident_module"
	"auto_park/modules/notification_module"
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

	user_module.RegisterRoutes(r, cfg, db)
	vehicle_module.RegisterRoutes(r, cfg, db)
	tripsheet_module.RegisterRoutes(r, cfg, db)
	incident_module.RegisterRoutes(r, cfg, db)
	notificationHub := notificationws.NewHub()
	notificationRepo := notificationrepository.NewNotificationRepository(db)
	notificationSvc := notificationservice.NewNotificationService(notificationRepo, notificationHub)
	notification_module.RegisterRoutes(r, cfg, notificationSvc, notificationHub)

	fuel_module.RegisterRoutes(r, cfg, db)
	warehouse_module.RegisterRoutes(r, cfg, db, notificationSvc)

	return r
}
